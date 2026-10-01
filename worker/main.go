package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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

	if cookies := os.Getenv("YOUTUBE_COOKIES"); strings.TrimSpace(cookies) != "" {
		if err := os.WriteFile("/tmp/youtube-cookies.txt", []byte(cookies), 0600); err != nil {
			log.Printf("failed to write youtube cookies file: %v", err)
		}
	}

	// SAFE COOKIE DIAGNOSTICS:
	hasEnv := strings.TrimSpace(os.Getenv("YOUTUBE_COOKIES")) != ""
	log.Printf("DIAG YOUTUBE_COOKIES env set: %t", hasEnv)
	if fi, err := os.Stat("/tmp/youtube-cookies.txt"); err == nil {
		log.Printf("DIAG /tmp/youtube-cookies.txt exists: true, size: %d bytes, mode: %s", fi.Size(), fi.Mode().String())
	} else {
		log.Printf("DIAG /tmp/youtube-cookies.txt exists: false (err: %v)", err)
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
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
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
