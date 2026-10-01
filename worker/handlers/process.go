package handlers

import (
	"net/http"

	"github.com/alldownload/worker/security"
)

type ProcessRequest struct {
	JobID string `json:"jobId"`
}

func (a *App) ProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.authenticate(w, r) {
		return
	}
	var req ProcessRequest
	if err := decodeJSON(w, r, a.cfg.MaxRequestBody, &req); err != nil || security.ValidateUUID(req.JobID) != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	job, err := a.store.GetJob(r.Context(), req.JobID)
	if err != nil {
		if isNotFound(err) {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Job service unavailable", http.StatusServiceUnavailable)
		return
	}
	if job.Status == "ready" || job.Status == "error" || job.Status == "expired" {
		http.Error(w, "Job is not processable", http.StatusConflict)
		return
	}
	if err := a.processor.Trigger(r.Context(), req.JobID); err != nil {
		http.Error(w, "Could not queue job", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}
