package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alldownload/worker/config"
	"github.com/alldownload/worker/platforms"
	"github.com/alldownload/worker/security"
)

type Processor struct {
	cfg       config.Config
	store     *Store
	semaphore chan struct{}
}

func NewProcessor(cfg config.Config, store *Store) *Processor {
	return &Processor{cfg: cfg, store: store, semaphore: make(chan struct{}, cfg.MaxConcurrentJobs)}
}

func (p *Processor) Start(ctx context.Context) {
	go p.poll(ctx)
	go p.cleanup(ctx)
}

func (p *Processor) Trigger(ctx context.Context, jobID string) error {
	if !p.tryAcquire() {
		return nil // The durable queue will pick this job up when capacity is available.
	}
	job, err := p.store.ClaimJob(ctx, jobID)
	if err != nil {
		p.release()
		if errors.Is(err, ErrNotClaimable) {
			return nil
		}
		return err
	}
	go p.runClaimed(context.WithoutCancel(ctx), job)
	return nil
}

func (p *Processor) poll(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.JobPollInterval)
	defer ticker.Stop()
	for {
		p.claimOne(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *Processor) claimOne(ctx context.Context) {
	if !p.tryAcquire() {
		return
	}
	job, err := p.store.ClaimNextJob(ctx)
	if err != nil {
		p.release()
		if !errors.Is(err, ErrNotClaimable) && !errors.Is(err, context.Canceled) {
			log.Printf("job claim failed: %v", err)
		}
		return
	}
	go p.runClaimed(ctx, job)
}

func (p *Processor) cleanup(ctx context.Context) {
	if err := p.store.CleanupExpired(ctx); err != nil {
		log.Printf("expired object cleanup failed: %v", err)
	}
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.store.CleanupExpired(ctx); err != nil {
				log.Printf("expired object cleanup failed: %v", err)
			}
		}
	}
}

func (p *Processor) runClaimed(parent context.Context, job *DownloadJob) {
	defer p.release()
	ctx, cancel := context.WithTimeout(parent, p.cfg.DownloadTimeout+p.cfg.UploadTimeout+p.cfg.ResolveTimeout)
	defer cancel()
	if err := p.process(ctx, job); err != nil {
		log.Printf("[JOB %s] failed: %v", job.ID, err)
		if ctx.Err() != nil {
			_ = p.store.PatchJob(context.Background(), job.ID, map[string]any{
				"status": "queued", "stage": "queued", "progress": nil,
				"locked_at": nil, "locked_by": nil,
			})
			return
		}
		if updateErr := p.store.UpdateJobError(context.Background(), job.ID, errorCode(err), safeErrorMessage(err)); updateErr != nil {
			log.Printf("[JOB %s] failed to persist error state: %v", job.ID, updateErr)
		}
	}
}

func (p *Processor) process(ctx context.Context, job *DownloadJob) error {
	if err := security.ValidateUUID(job.ID); err != nil {
		return err
	}
	if !security.ValidateOutput(job.OutputFormat) || !security.ValidateFormatID(job.SelectedFormat) {
		return fmt.Errorf("invalid job format (output=%q, selected=%q)", job.OutputFormat, job.SelectedFormat)
	}
	if job.MediaType == "video" && job.OutputFormat != "mp4" && job.OutputFormat != "mp3" {
		return fmt.Errorf("invalid video output")
	}
	if job.MediaType == "image" && job.OutputFormat != "jpg" && job.OutputFormat != "png" && job.OutputFormat != "webp" {
		return fmt.Errorf("invalid image output")
	}

	media, err := p.store.GetResolvedMedia(ctx, job.ResolvedMediaID)
	if err != nil {
		return fmt.Errorf("resolved media unavailable")
	}
	if media.ExpiresAt != nil && time.Now().After(*media.ExpiresAt) {
		return fmt.Errorf("resolved media expired")
	}
	if media.SourceURL != job.SourceURL || media.Platform != job.Platform || media.MediaType != job.MediaType {
		return fmt.Errorf("job metadata does not match resolved media")
	}
	if !formatExists(media.Formats, job.SelectedFormat) {
		return fmt.Errorf("selected format is not available")
	}

	leaseCtx, stopLease := context.WithCancel(ctx)
	defer stopLease()
	go p.renewLease(leaseCtx, job.ID)

	if err := p.store.UpdateJobState(ctx, job.ID, "processing", "extracting", nil); err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp(os.TempDir(), "media-job-")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	if err := os.Chmod(tempDir, 0700); err != nil {
		return fmt.Errorf("secure temporary directory: %w", err)
	}

	adapter := platforms.GetAdapter(job.Platform)
	if adapter == nil {
		return fmt.Errorf("unsupported platform")
	}
	if err := p.store.UpdateJobState(ctx, job.ID, "processing", "downloading", nil); err != nil {
		return err
	}
	resultFile, err := adapter.Download(ctx, job.ID, job.SourceURL, job.SelectedFormat, job.OutputFormat, tempDir)
	if err != nil {
		return fmt.Errorf("download failed")
	}
	fileInfo, err := os.Stat(resultFile)
	if err != nil || !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("downloaded file is unavailable")
	}
	if fileInfo.Size() <= 0 || fileInfo.Size() > p.cfg.MaxFileSize {
		return fmt.Errorf("downloaded file exceeds configured size limit")
	}

	if err := p.store.UpdateJobState(ctx, job.ID, "processing", "uploading", nil); err != nil {
		return err
	}
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(resultFile)), ".")
	filename := security.SafeFilename(job.Title, extension)
	storagePath := "downloads/" + job.ID + "/" + filename
	if err := p.store.UploadToStorage(ctx, resultFile, storagePath); err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}
	if err := p.store.UpdateJobReady(ctx, job.ID, storagePath, filename, fileInfo.Size()); err != nil {
		return err
	}
	log.Printf("[JOB %s] ready platform=%s bytes=%d", job.ID, job.Platform, fileInfo.Size())
	return nil
}

func (p *Processor) renewLease(ctx context.Context, jobID string) {
	interval := p.cfg.JobLeaseTimeout / 3
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.store.PatchJob(ctx, jobID, map[string]any{
				"locked_at": time.Now().UTC().Format(time.RFC3339), "locked_by": p.cfg.WorkerID,
			}); err != nil {
				log.Printf("[JOB %s] lease renewal failed: %v", jobID, err)
			}
		}
	}
}

func (p *Processor) tryAcquire() bool {
	select {
	case p.semaphore <- struct{}{}:
		return true
	default:
		return false
	}
}

func (p *Processor) release() { <-p.semaphore }

func formatExists(formats []platforms.MediaFormat, selected string) bool {
	for _, format := range formats {
		if format.ID == selected {
			return true
		}
	}
	return false
}

func errorCode(err error) string {
	message := err.Error()
	if strings.Contains(message, "upload") {
		return "upload_failed"
	}
	if strings.Contains(message, "format") {
		return "invalid_format"
	}
	return "processing_failed"
}

func safeErrorMessage(err error) string {
	message := err.Error()
	if len(message) > 160 {
		return message[:160]
	}
	return message
}
