package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Result represents the raw JSON output from yt-dlp --dump-single-json
type Result struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Thumbnail   string   `json:"thumbnail"`
	Duration    float64  `json:"duration"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Formats     []Format `json:"formats"`
	URL         string   `json:"url"`
	Ext         string   `json:"ext"`
	Extractor   string   `json:"extractor"`
	WebpageURL  string   `json:"webpage_url"`
	Description string   `json:"description"`
}

// Format represents a single format from yt-dlp
type Format struct {
	FormatID       string  `json:"format_id"`
	Ext            string  `json:"ext"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	FPS            float64 `json:"fps"`
	VCodec         string  `json:"vcodec"`
	ACodec         string  `json:"acodec"`
	Filesize       int64   `json:"filesize"`
	FilesizeApprox int64   `json:"filesize_approx"`
	FormatNote     string  `json:"format_note"`
	URL            string  `json:"url"`
	TBR            float64 `json:"tbr"`
}

// Resolve runs yt-dlp --dump-single-json and returns parsed output
func Resolve(ctx context.Context, url string) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--no-playlist",
		"--dump-single-json",
		"--no-download",
		"--no-warnings",
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := string(output)
		if len(message) > 1200 {
			message = message[len(message)-1200:]
		}
		return nil, fmt.Errorf("yt-dlp resolve failed: %s: %w", message, err)
	}

	// yt-dlp can print runtime notices before the JSON payload. For example,
	// Python 3.10 currently emits a deprecation notice on stdout. Strip that
	// preamble before decoding the actual JSON document.
	start := bytes.IndexByte(output, '{')
	end := bytes.LastIndexByte(output, '}')
	if start < 0 || end < start {
		return nil, fmt.Errorf("yt-dlp returned no JSON metadata")
	}

	var result Result
	if err := json.Unmarshal(output[start:end+1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	return &result, nil
}

// Download downloads media using yt-dlp with the given format
func Download(ctx context.Context, url, formatSpec, outputPath string, extraArgs ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	args := []string{
		"--no-playlist",
		"--no-warnings",
	}

	if formatSpec != "" {
		args = append(args, "-f", formatSpec)
	}

	args = append(args, "-o", outputPath)
	args = append(args, extraArgs...)
	args = append(args, url)

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp download failed: %s: %w", string(out), err)
	}

	return nil
}
