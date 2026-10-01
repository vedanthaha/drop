package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/alldownload/worker/platforms"
	"github.com/alldownload/worker/security"
	"github.com/alldownload/worker/services"
)

type ResolveRequest struct {
	URL string `json:"url"`
}

func (a *App) ResolveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.authenticate(w, r) {
		return
	}
	var req ResolveRequest
	if err := decodeJSON(w, r, a.cfg.MaxRequestBody, &req); err != nil || req.URL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	platform, err := security.ValidateMediaURL(req.URL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported or invalid media URL"})
		return
	}
	adapter := platforms.GetAdapter(platform)
	if adapter == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported platform"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), a.cfg.ResolveTimeout)
	defer cancel()
	result, err := adapter.Resolve(ctx, req.URL)
	if err != nil {
		log.Printf("Resolve failed for %s: %v", req.URL, err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Could not resolve this link"})
		return
	}
	result.SourceURL = req.URL
	result.ID = security.GenerateUUID()
	expiresAt := time.Now().UTC().Add(a.cfg.JobRetention)

	resolvedRecord := &services.ResolvedMedia{
		ID:        result.ID,
		SourceURL: result.SourceURL,
		Platform:  result.Platform,
		MediaType: result.Type,
		Title:     result.Title,
		Thumbnail: result.Thumbnail,
		Width:     result.Width,
		Height:    result.Height,
		Duration:  result.Duration,
		Formats:   result.Formats,
		ExpiresAt: &expiresAt,
	}

	if err := a.store.SaveResolvedMedia(ctx, resolvedRecord); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not persist resolved media"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
