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
	http.HandleFunc("/resolve", handlers.ResolveHandler)
	http.HandleFunc("/process", handlers.ProcessHandler)
	http.HandleFunc("/jobs/", handlers.JobsHandler)

	log.Printf("Worker starting on port %s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
