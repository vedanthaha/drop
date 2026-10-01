const fs = require('fs');
const path = require('path');

const files = {
"worker/main.go": `
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/alldownload/worker/handlers"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() // ignore error if .env doesn't exist

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/process", handlers.ProcessHandler)
	http.HandleFunc("/jobs/", handlers.JobsHandler)

	log.Printf("Worker starting on port %s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
`,

"worker/handlers/health.go": `
package handlers

import (
	"encoding/json"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
`,

"worker/handlers/process.go": `
package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/alldownload/worker/services"
)

type ProcessRequest struct {
	JobID     string \`json:"jobId"\`
	SourceURL string \`json:"sourceUrl"\`
	Platform  string \`json:"platform"\`
	FormatID  string \`json:"formatId"\`
	Output    string \`json:"output"\`
}

func ProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	secret := os.Getenv("WORKER_SECRET")
	if secret != "" && r.Header.Get("Authorization") != "Bearer "+secret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Run process asynchronously
	go services.ProcessJob(context.Background(), req.JobID, req.SourceURL, req.Platform, req.FormatID, req.Output)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}
`,

"worker/handlers/jobs.go": `
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alldownload/worker/services"
)

func JobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	jobID := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	job, err := services.GetJobStatus(context.Background(), jobID)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
`,

"worker/services/storage.go": `
package services

import (
	"context"
	"fmt"
	"os"

	"github.com/supabase-community/supabase-go"
)

var client *supabase.Client

func init() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL != "" && supabaseKey != "" {
		c, err := supabase.NewClient(supabaseURL, supabaseKey, &supabase.ClientOptions{})
		if err == nil {
			client = c
		}
	}
}

type JobState struct {
	Status   string \`json:"status"\`
	Stage    string \`json:"stage"\`
	Progress int    \`json:"progress"\`
}

func GetJobStatus(ctx context.Context, jobID string) (*JobState, error) {
	// Stub implementation for now - normally queries Supabase DB
	if client == nil {
		return nil, fmt.Errorf("supabase client not initialized")
	}
	// TODO: read from DB
	return &JobState{Status: "unknown", Stage: "unknown"}, nil
}

func UpdateJobState(jobID, status, stage string, progress int) {
    if client == nil { return }
	// Update row in DB via REST or Postgres directly
	// e.g. client.DB.From("download_jobs").Update(map[string]interface{}{"status": status, "stage": stage, "progress": progress}).Eq("id", jobID).Execute()
}

func UploadToStorage(jobID, localFilePath, destinationPath string) error {
	if client == nil {
		return fmt.Errorf("supabase client not initialized")
	}
	file, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Perform upload using Supabase storage client (stub - community client storage might require direct HTTP call or storage api)
	// For simplicity, we assume there's a storage method or we make a direct HTTP POST request to Supabase Storage API.
	return nil
}
`,

"worker/services/downloader.go": `
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alldownload/worker/platforms"
)

func ProcessJob(ctx context.Context, jobID, url, platform, formatID, output string) {
	UpdateJobState(jobID, "processing", "queued", 0)

	// Create temp directory for job
	tempDir := filepath.Join(os.TempDir(), "media-jobs", jobID)
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir) // cleanup

	UpdateJobState(jobID, "processing", "downloading", 10)

	adapter := platforms.GetAdapter(platform)
	if adapter == nil {
		UpdateJobState(jobID, "error", "failed", 0)
		return
	}

	resultFile, err := adapter.Download(ctx, jobID, url, formatID, output, tempDir)
	if err != nil {
		fmt.Printf("Download error for job %s: %v\n", jobID, err)
		UpdateJobState(jobID, "error", "failed", 0)
		return
	}

	UpdateJobState(jobID, "processing", "uploading", 90)
	
	// Upload
	destPath := fmt.Sprintf("downloads/%s/%s", jobID, filepath.Base(resultFile))
	err = UploadToStorage(jobID, resultFile, destPath)
	if err != nil {
		fmt.Printf("Upload error for job %s: %v\n", jobID, err)
		UpdateJobState(jobID, "error", "failed", 0)
		return
	}

	UpdateJobState(jobID, "ready", "ready", 100)
}
`,

"worker/platforms/base.go": `
package platforms

import "context"

type MediaFormat struct {
	ID       string \`json:"id"\`
	Label    string \`json:"label"\`
	Width    int    \`json:"width"\`
	Height   int    \`json:"height"\`
	Ext      string \`json:"ext"\`
	HasAudio bool   \`json:"hasAudio"\`
	Filesize int64  \`json:"filesize"\`
}

type MediaResult struct {
	ID        string        \`json:"id"\`
	Platform  string        \`json:"platform"\`
	Type      string        \`json:"type"\`
	Title     string        \`json:"title"\`
	Thumbnail string        \`json:"thumbnail"\`
	Width     int           \`json:"width"\`
	Height    int           \`json:"height"\`
	Duration  int           \`json:"duration"\`
	Formats   []MediaFormat \`json:"formats"\`
}

type Adapter interface {
	CanHandle(url string) bool
	Resolve(ctx context.Context, url string) (*MediaResult, error)
	Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error)
}
`,

"worker/platforms/detector.go": `
package platforms

var adapters = map[string]Adapter{
	"youtube":   &YouTubeAdapter{},
	"pinterest": &PinterestAdapter{},
}

func GetAdapter(platform string) Adapter {
	return adapters[platform]
}

func DetectPlatform(url string) string {
	for name, adapter := range adapters {
		if adapter.CanHandle(url) {
			return name
		}
	}
	return "unknown"
}
`,

"worker/platforms/youtube.go": `
package platforms

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type YouTubeAdapter struct{}

func (y *YouTubeAdapter) CanHandle(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

func (y *YouTubeAdapter) Resolve(ctx context.Context, url string) (*MediaResult, error) {
	// Stub implementation
	// Real implementation uses exec.CommandContext(ctx, "yt-dlp", "--dump-single-json", url)
	return &MediaResult{}, nil
}

func (y *YouTubeAdapter) Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error) {
	outPath := filepath.Join(tempDir, fmt.Sprintf("video.%s", output))
	
	// yt-dlp -f formatID -o outPath url
	cmd := exec.CommandContext(ctx, "yt-dlp", "-f", formatID, "-o", outPath, url)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outPath, nil
}
`,

"worker/platforms/pinterest.go": `
package platforms

import (
	"context"
	"strings"
)

type PinterestAdapter struct{}

func (p *PinterestAdapter) CanHandle(url string) bool {
	return strings.Contains(url, "pinterest.com") || strings.Contains(url, "pin.it")
}

func (p *PinterestAdapter) Resolve(ctx context.Context, url string) (*MediaResult, error) {
	return &MediaResult{}, nil
}

func (p *PinterestAdapter) Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error) {
	// Use upscale processor if output format specifies it
	return "", nil
}
`
};

for (const [filepath, content] of Object.entries(files)) {
  const fullPath = path.join(__dirname, filepath);
  fs.mkdirSync(path.dirname(fullPath), { recursive: true });
  fs.writeFileSync(fullPath, content.trim() + '\n');
}
console.log('Worker files generated successfully.');
