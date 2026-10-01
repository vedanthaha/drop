package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/alldownload/worker/services"
)

type ProcessRequest struct {
	JobID     string `json:"jobId"`
	SourceURL string `json:"sourceUrl"`
	Platform  string `json:"platform"`
	FormatID  string `json:"formatId"`
	Output    string `json:"output"`
	Title     string `json:"title"`
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
	go services.ProcessJob(context.Background(), req.JobID, req.SourceURL, req.Platform, req.FormatID, req.Output, req.Title)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}
