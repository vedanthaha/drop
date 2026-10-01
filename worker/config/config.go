package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                    string
	SupabaseURL             string
	SupabaseServiceRoleKey string
	WorkerSecret            string
	WorkerID                string
	ResolveTimeout          time.Duration
	DownloadTimeout         time.Duration
	UploadTimeout           time.Duration
	JobLeaseTimeout         time.Duration
	JobPollInterval         time.Duration
	JobRetention            time.Duration
	MaxConcurrentJobs       int
	MaxFileSize             int64
	MaxRequestBody          int64
}

func Load() (Config, error) {
	cfg := Config{
		Port:              envOr("PORT", "8080"),
		SupabaseURL:       os.Getenv("SUPABASE_URL"),
		WorkerSecret:      os.Getenv("WORKER_SECRET"),
		WorkerID:          envOr("WORKER_ID", ""),
		ResolveTimeout:    durationOr("RESOLVE_TIMEOUT", 90*time.Second),
		DownloadTimeout:   durationOr("DOWNLOAD_TIMEOUT", 10*time.Minute),
		UploadTimeout:     durationOr("UPLOAD_TIMEOUT", 10*time.Minute),
		JobLeaseTimeout:   durationOr("JOB_LEASE_TIMEOUT", 15*time.Minute),
		JobPollInterval:   durationOr("JOB_POLL_INTERVAL", 5*time.Second),
		JobRetention:      durationOr("JOB_RETENTION", 60*time.Minute),
		MaxConcurrentJobs: intOr("MAX_CONCURRENT_JOBS", 1),
		MaxFileSize:       int64Or("MAX_FILE_SIZE", 50*1024*1024),
		MaxRequestBody:    int64Or("MAX_REQUEST_BODY", 16*1024),
	}

	cfg.SupabaseServiceRoleKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if cfg.SupabaseURL == "" || cfg.SupabaseServiceRoleKey == "" || cfg.WorkerSecret == "" {
		return Config{}, fmt.Errorf("SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY, and WORKER_SECRET are required")
	}
	parsed, err := url.Parse(cfg.SupabaseURL)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return Config{}, fmt.Errorf("SUPABASE_URL must be a valid https URL")
	}
	if len(cfg.WorkerSecret) < 16 {
		return Config{}, fmt.Errorf("WORKER_SECRET must be at least 16 characters")
	}
	if cfg.MaxConcurrentJobs < 1 || cfg.MaxFileSize < 1 || cfg.MaxRequestBody < 1 {
		return Config{}, fmt.Errorf("worker limits must be positive")
	}
	if cfg.WorkerID == "" {
		cfg.WorkerID = envOr("HOSTNAME", "worker")
	}
	return cfg, nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func durationOr(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func intOr(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func int64Or(name string, fallback int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
