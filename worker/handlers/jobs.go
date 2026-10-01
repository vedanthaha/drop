package handlers

import (
	"context"
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
