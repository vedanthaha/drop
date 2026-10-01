package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alldownload/worker/config"
	"github.com/alldownload/worker/platforms"
)

var ErrNotFound = errors.New("not found")
var ErrNotClaimable = errors.New("job is not claimable")

type Store struct {
	cfg        config.Config
	httpClient *http.Client
}

type DownloadJob struct {
	ID              string                    `json:"id"`
	SourceURL       string                    `json:"source_url"`
	Platform        string                    `json:"platform"`
	MediaType       string                    `json:"media_type"`
	Status          string                    `json:"status"`
	Stage           string                    `json:"stage"`
	Title           string                    `json:"title"`
	SelectedFormat  string                    `json:"selected_format"`
	OutputFormat    string                    `json:"output_format"`
	ResolvedMediaID string                    `json:"resolved_media_id"`
	StoragePath     string                    `json:"storage_path"`
	Filename        string                    `json:"filename"`
	Filesize        int64                     `json:"filesize"`
	Progress        *int                      `json:"progress"`
	ErrorCode       string                    `json:"error_code"`
	ErrorMessage    string                    `json:"error_message"`
	RetryCount      int                       `json:"retry_count"`
	ExpiresAt       *time.Time                `json:"expires_at"`
	Formats         []platforms.MediaFormat   `json:"formats,omitempty"`
}

type ResolvedMedia struct {
	ID        string                  `json:"id"`
	SourceURL string                  `json:"source_url"`
	Platform  string                  `json:"platform"`
	MediaType string                  `json:"media_type"`
	Title     string                  `json:"title"`
	Thumbnail string                  `json:"thumbnail_url"`
	Width     int                     `json:"width"`
	Height    int                     `json:"height"`
	Duration  int                     `json:"duration"`
	Formats   []platforms.MediaFormat `json:"formats"`
	ExpiresAt *time.Time              `json:"expires_at"`
}

type expiredObject struct {
	ID          string `json:"id"`
	StoragePath string `json:"storage_path"`
}

func NewStore(cfg config.Config) *Store {
	return &Store{
		cfg: cfg,
		httpClient: &http.Client{Timeout: cfg.UploadTimeout},
	}
}

func (s *Store) GetJobStatus(ctx context.Context, jobID string) (*DownloadJob, error) {
	values := url.Values{"select": {"status,stage,progress,error_message,filename,filesize"}, "id": {"eq." + jobID}}
	var jobs []DownloadJob
	if err := s.getJSON(ctx, "/rest/v1/download_jobs?"+values.Encode(), &jobs); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, ErrNotFound
	}
	return &jobs[0], nil
}

func (s *Store) GetJob(ctx context.Context, jobID string) (*DownloadJob, error) {
	values := url.Values{"select": {"id,source_url,platform,media_type,status,stage,title,selected_format,output_format,resolved_media_id,storage_path,filename,filesize,progress,error_code,error_message,retry_count,expires_at"}, "id": {"eq." + jobID}}
	var jobs []DownloadJob
	if err := s.getJSON(ctx, "/rest/v1/download_jobs?"+values.Encode(), &jobs); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, ErrNotFound
	}
	return &jobs[0], nil
}

func (s *Store) GetResolvedMedia(ctx context.Context, mediaID string) (*ResolvedMedia, error) {
	values := url.Values{"select": {"id,source_url,platform,media_type,title,thumbnail_url,width,height,duration,formats,expires_at"}, "id": {"eq." + mediaID}}
	var media []ResolvedMedia
	if err := s.getJSON(ctx, "/rest/v1/resolved_media?"+values.Encode(), &media); err != nil {
		return nil, err
	}
	if len(media) == 0 {
		return nil, ErrNotFound
	}
	return &media[0], nil
}

func (s *Store) SaveResolvedMedia(ctx context.Context, media *ResolvedMedia) error {
	var target []ResolvedMedia
	return s.postJSON(ctx, "/rest/v1/resolved_media", media, &target)
}

func (s *Store) ClaimJob(ctx context.Context, jobID string) (*DownloadJob, error) {
	payload := map[string]any{
		"p_job_id":       jobID,
		"p_worker_id":    s.cfg.WorkerID,
		"p_lease_seconds": int(s.cfg.JobLeaseTimeout.Seconds()),
	}
	var jobs []DownloadJob
	if err := s.postJSON(ctx, "/rest/v1/rpc/claim_download_job", payload, &jobs); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, ErrNotClaimable
	}
	return &jobs[0], nil
}

func (s *Store) ClaimNextJob(ctx context.Context) (*DownloadJob, error) {
	payload := map[string]any{
		"p_worker_id":    s.cfg.WorkerID,
		"p_lease_seconds": int(s.cfg.JobLeaseTimeout.Seconds()),
	}
	var jobs []DownloadJob
	if err := s.postJSON(ctx, "/rest/v1/rpc/claim_next_download_job", payload, &jobs); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, ErrNotClaimable
	}
	return &jobs[0], nil
}

func (s *Store) PatchJob(ctx context.Context, jobID string, payload map[string]any) error {
	values := url.Values{"id": {"eq." + jobID}}
	return s.patchJSON(ctx, "/rest/v1/download_jobs?"+values.Encode(), payload)
}

func (s *Store) UpdateJobState(ctx context.Context, jobID, status, stage string, progress *int) error {
	return s.PatchJob(ctx, jobID, map[string]any{
		"status": status, "stage": stage, "progress": progress, "updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Store) UpdateJobError(ctx context.Context, jobID, code, message string) error {
	return s.PatchJob(ctx, jobID, map[string]any{
		"status": "error", "stage": "failed", "progress": nil,
		"error_code": code, "error_message": message, "updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Store) UpdateJobReady(ctx context.Context, jobID, storagePath, filename string, filesize int64) error {
	return s.PatchJob(ctx, jobID, map[string]any{
		"status": "ready", "stage": "ready", "progress": 100,
		"storage_path": storagePath, "filename": filename, "filesize": filesize,
		"expires_at": time.Now().UTC().Add(s.cfg.JobRetention).Format(time.RFC3339),
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Store) UploadToStorage(ctx context.Context, localFilePath, destinationPath string) error {
	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		return err
	}
	if fileInfo.Size() > s.cfg.MaxFileSize {
		return fmt.Errorf("file exceeds configured size limit")
	}

	for attempt := 0; attempt < 3; attempt++ {
		file, openErr := os.Open(localFilePath)
		if openErr != nil {
			return openErr
		}
		endpoint := s.storageURL(destinationPath)
		req, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, file)
		if requestErr != nil {
			file.Close()
			return requestErr
		}
		req.ContentLength = fileInfo.Size()
		req.Header.Set("apikey", s.cfg.SupabaseServiceRoleKey)
		req.Header.Set("Authorization", "Bearer "+s.cfg.SupabaseServiceRoleKey)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Upsert", "false")
		resp, requestErr := s.httpClient.Do(req)
		if requestErr == nil {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			requestErr = fmt.Errorf("storage upload returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		file.Close()
		if attempt == 2 {
			return requestErr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
		}
	}
	return fmt.Errorf("storage upload failed")
}

func (s *Store) DeleteStorageObject(ctx context.Context, storagePath string) error {
	if storagePath == "" {
		return nil
	}
	return s.do(ctx, http.MethodDelete, s.storageURL(storagePath), nil, nil)
}

func (s *Store) CleanupExpired(ctx context.Context) error {
	values := url.Values{"select": {"id,storage_path"}, "status": {"in.(ready,error)"}, "expires_at": {"lt." + time.Now().UTC().Format(time.RFC3339)}}
	var objects []expiredObject
	if err := s.getJSON(ctx, "/rest/v1/download_jobs?"+values.Encode(), &objects); err != nil {
		return err
	}
	for _, object := range objects {
		if err := s.DeleteStorageObject(ctx, object.StoragePath); err != nil {
			continue
		}
		_ = s.PatchJob(ctx, object.ID, map[string]any{"status": "expired", "stage": "expired", "storage_path": nil, "updated_at": time.Now().UTC().Format(time.RFC3339)})
	}
	return nil
}

func (s *Store) storageURL(storagePath string) string {
	return strings.TrimRight(s.cfg.SupabaseURL, "/") + "/storage/v1/object/media/" + path.Join(strings.TrimPrefix(storagePath, "/"))
}

func (s *Store) getJSON(ctx context.Context, endpoint string, target any) error {
	body, err := s.request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func (s *Store) postJSON(ctx context.Context, endpoint string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	response, err := s.request(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(response, target)
}

func (s *Store) patchJSON(ctx context.Context, endpoint string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.request(ctx, http.MethodPatch, endpoint, body)
	return err
}

func (s *Store) do(ctx context.Context, method, endpoint string, body []byte, target any) error {
	response, err := s.request(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	if target != nil {
		return json.Unmarshal(response, target)
	}
	return nil
}

func (s *Store) request(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(s.cfg.SupabaseURL, "/")+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.cfg.SupabaseServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+s.cfg.SupabaseServiceRoleKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Prefer", "return=representation")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Supabase request returned %d", resp.StatusCode)
	}
	return responseBody, nil
}
