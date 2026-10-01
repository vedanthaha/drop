package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/alldownload/worker/platforms"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 \-_]+`)

func cleanTitle(title string) string {
	if title == "" {
		return "media"
	}
	// Replace invalid characters with nothing, but preserve spaces and dashes
	cleaned := nonAlphanumericRegex.ReplaceAllString(title, "")
	return cleaned
}

func ProcessJob(ctx context.Context, jobID, url, platform, formatID, output, title string) {
	UpdateJobState(jobID, "processing", "extracting", 0)

	// Create temp directory for job
	tempDir := filepath.Join(os.TempDir(), "media-jobs", jobID)
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir) // always cleanup

	adapter := platforms.GetAdapter(platform)
	if adapter == nil {
		UpdateJobError(jobID, "unsupported_platform", "Platform not supported")
		return
	}

	UpdateJobState(jobID, "processing", "downloading", 20)

	resultFile, err := adapter.Download(ctx, jobID, url, formatID, output, tempDir)
	if err != nil {
		fmt.Printf("Download error for job %s: %v\n", jobID, err)
		UpdateJobError(jobID, "download_failed", "Download failed")
		return
	}

	// Get file info
	fi, err := os.Stat(resultFile)
	if err != nil {
		fmt.Printf("Stat error for job %s: %v\n", jobID, err)
		UpdateJobError(jobID, "file_error", "Could not read downloaded file")
		return
	}

	UpdateJobState(jobID, "processing", "uploading", 80)

	// Upload to Supabase Storage
	ext := filepath.Ext(resultFile)
	finalFilename := fmt.Sprintf("%s%s", cleanTitle(title), ext)
	
	destPath := fmt.Sprintf("downloads/%s/%s", jobID, finalFilename)
	err = UploadToStorage(jobID, resultFile, destPath)
	if err != nil {
		fmt.Printf("Upload error for job %s: %v\n", jobID, err)
		UpdateJobError(jobID, "upload_failed", "Upload failed")
		return
	}

	// Mark job as ready with file metadata
	UpdateJobReady(jobID, destPath, finalFilename, fi.Size())
	fmt.Printf("Job %s completed successfully: %s (%d bytes)\n", jobID, finalFilename, fi.Size())
}

func UpdateJobError(jobID, code, message string) {
	if supabaseURL == "" || supabaseKey == "" {
		return
	}

	payload := fmt.Sprintf(`{"status":"error","error_code":"%s","error_message":"%s"}`, code, message)
	patchJob(jobID, []byte(payload))
}

func UpdateJobReady(jobID, storagePath, filename string, filesize int64) {
	if supabaseURL == "" || supabaseKey == "" {
		return
	}

	expiresAt := time.Now().Add(60 * time.Minute).UTC().Format(time.RFC3339)
	payload := fmt.Sprintf(
		`{"status":"ready","stage":"ready","progress":100,"storage_path":"%s","filename":"%s","filesize":%d,"expires_at":"%s"}`,
		storagePath, filename, filesize, expiresAt,
	)
	patchJob(jobID, []byte(payload))
}
