package handlers

import (
	"net/http"
	"strings"

	"github.com/alldownload/worker/security"
)

func (a *App) JobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.authenticate(w, r) {
		return
	}
	jobID := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if security.ValidateUUID(jobID) != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}
	job, err := a.store.GetJobStatus(r.Context(), jobID)
	if err != nil {
		if isNotFound(err) {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Job service unavailable", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, job)
}
