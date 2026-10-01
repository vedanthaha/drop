package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	supabaseURL string
	supabaseKey string
	httpClient  = &http.Client{Timeout: 120 * time.Second}
)

func init() {
	supabaseURL = os.Getenv("SUPABASE_URL")
	supabaseKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
}

type JobState struct {
	Status   string `json:"status"`
	Stage    string `json:"stage"`
	Progress int    `json:"progress"`
}

func GetJobStatus(ctx context.Context, jobID string) (*JobState, error) {
	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("supabase not configured")
	}

	url := fmt.Sprintf("%s/rest/v1/download_jobs?id=eq.%s&select=status,stage,progress", supabaseURL, jobID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []JobState
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("job not found")
	}
	return &results[0], nil
}

func UpdateJobState(jobID, status, stage string, progress int) {
	if supabaseURL == "" || supabaseKey == "" {
		return
	}

	payload := map[string]interface{}{
		"status":   status,
		"stage":    stage,
		"progress": progress,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/rest/v1/download_jobs?id=eq.%s", supabaseURL, jobID)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		fmt.Printf("UpdateJobState error: %v\n", err)
		return
	}
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("UpdateJobState request error: %v\n", err)
		return
	}
	resp.Body.Close()
}

func UploadToStorage(jobID, localFilePath, destinationPath string) error {
	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("supabase not configured")
	}

	file, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	url := fmt.Sprintf("%s/storage/v1/object/media/%s", supabaseURL, destinationPath)
	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func patchJob(jobID string, body []byte) {
	url := fmt.Sprintf("%s/rest/v1/download_jobs?id=eq.%s", supabaseURL, jobID)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		fmt.Printf("patchJob error: %v\n", err)
		return
	}
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("patchJob request error: %v\n", err)
		return
	}
	resp.Body.Close()
}
