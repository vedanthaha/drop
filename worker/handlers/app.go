package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alldownload/worker/config"
	"github.com/alldownload/worker/services"
	"github.com/alldownload/worker/security"
)

type App struct {
	cfg       config.Config
	store     *services.Store
	processor *services.Processor
}

func NewApp(cfg config.Config, store *services.Store, processor *services.Processor) *App {
	return &App{cfg: cfg, store: store, processor: processor}
}

func (a *App) authenticate(w http.ResponseWriter, r *http.Request) bool {
	if security.AuthenticateBearer(r, a.cfg.WorkerSecret) {
		return true
	}
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return false
}

func decodeJSON(w http.ResponseWriter, r *http.Request, maxBytes int64, target any) error {
	if r.ContentLength > maxBytes {
		return fmt.Errorf("request body too large")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("request must contain one JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func isNotFound(err error) bool { return errors.Is(err, services.ErrNotFound) }
