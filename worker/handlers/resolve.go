package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/alldownload/worker/platforms"
)

type ResolveRequest struct {
	URL string `json:"url"`
}

func ResolveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	secret := os.Getenv("WORKER_SECRET")
	if secret != "" && r.Header.Get("Authorization") != "Bearer "+secret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Detect platform
	platform := platforms.DetectPlatform(req.URL)
	if platform == "unknown" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unsupported platform"})
		return
	}

	adapter := platforms.GetAdapter(platform)
	if adapter == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Platform not supported"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	result, err := adapter.Resolve(ctx, req.URL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to resolve media"})
		return
	}

	result.SourceURL = req.URL

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
