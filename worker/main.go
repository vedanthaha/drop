package main

import (
	"context"
	"log"
	"net/http"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/alldownload/worker/config"
	"github.com/alldownload/worker/handlers"
	"github.com/alldownload/worker/services"
	"github.com/joho/godotenv"
)

func main() {
	// Load local development values before any application dependencies are built.
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	// DIAGNOSTICS:
	out, _ := exec.Command("yt-dlp", "--version").CombinedOutput()
	log.Printf("DIAG yt-dlp version: %s", string(out))
	out, _ = exec.Command("ls", "-la", "/usr/local/bin/deno").CombinedOutput()
	log.Printf("DIAG deno path: %s", string(out))
	out, _ = exec.Command("ls", "-la", "/app/bgutil").CombinedOutput()
	log.Printf("DIAG bgutil dir: %s", string(out))


	store := services.NewStore(cfg)
	processor := services.NewProcessor(cfg, store)
	app := handlers.NewApp(cfg, store, processor)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	processor.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.HealthHandler)
	mux.HandleFunc("/resolve", app.ResolveHandler)
	mux.HandleFunc("/process", app.ProcessHandler)
	mux.HandleFunc("/jobs/", app.JobsHandler)

	server := &http.Server{
		Addr:              "0.0.0.0:" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 * 1024,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("worker starting on port %s id=%s", cfg.Port, cfg.WorkerID)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
