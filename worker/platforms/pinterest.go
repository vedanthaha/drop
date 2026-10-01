package platforms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alldownload/worker/security"
	"github.com/alldownload/worker/ytdlp"
)

type PinterestAdapter struct{}

func (p *PinterestAdapter) CanHandle(url string) bool {
	platform, err := security.PlatformForURL(url)
	return err == nil && platform == "pinterest"
}

func (p *PinterestAdapter) Resolve(ctx context.Context, url string) (*MediaResult, error) {
	raw, err := ytdlp.Resolve(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("could not resolve Pinterest content: %w", err)
	}

	hasVideo := false
	for _, f := range raw.Formats {
		if f.VCodec != "none" && f.VCodec != "" {
			hasVideo = true
			break
		}
	}

	if hasVideo {
		type formatEntry struct {
			formatID string
			height   int
			width    int
			ext      string
			filesize int64
			tbr      float64
		}

		bestByHeight := make(map[int]*formatEntry)
		for _, f := range raw.Formats {
			if f.VCodec == "none" || f.VCodec == "" || f.Height == 0 {
				continue
			}
			fs := f.Filesize
			if fs == 0 {
				fs = f.FilesizeApprox
			}
			existing, ok := bestByHeight[f.Height]
			if !ok || f.TBR > existing.tbr {
				bestByHeight[f.Height] = &formatEntry{
					formatID: f.FormatID,
					height:   f.Height,
					width:    f.Width,
					ext:      f.Ext,
					filesize: fs,
					tbr:      f.TBR,
				}
			}
		}

		var heights []int
		for h := range bestByHeight {
			heights = append(heights, h)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(heights)))

		var formats []MediaFormat
		for _, h := range heights {
			e := bestByHeight[h]
			formats = append(formats, MediaFormat{
				ID:       e.formatID,
				Label:    HeightToLabel(e.height),
				Width:    e.width,
				Height:   e.height,
				Ext:      e.ext,
				HasAudio: true,
				Filesize: e.filesize,
			})
		}

		return &MediaResult{
			ID:        raw.ID,
			Platform:  "pinterest",
			Type:      "video",
			Title:     raw.Title,
			Thumbnail: raw.Thumbnail,
			Width:     raw.Width,
			Height:    raw.Height,
			Duration:  int(raw.Duration),
			Formats:   formats,
		}, nil
	}

	// Image pin
	width := raw.Width
	height := raw.Height
	formats := []MediaFormat{
		{
			ID:    "original",
			Label: fmt.Sprintf("Original %d\u00d7%d", width, height),
			Width: width, Height: height,
			Ext: "jpg",
		},
	}

	return &MediaResult{
		ID:        raw.ID,
		Platform:  "pinterest",
		Type:      "image",
		Title:     raw.Title,
		Thumbnail: raw.Thumbnail,
		Width:     width,
		Height:    height,
		Formats:   formats,
	}, nil
}

func (p *PinterestAdapter) Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error) {
	if formatID == "original" {
		return p.downloadImage(ctx, url, tempDir)
	}

	outPath := filepath.Join(tempDir, fmt.Sprintf("media.%s", output))
	var err error
	if output == "mp3" {
		err = ytdlp.Download(ctx, url, "bestaudio/best", outPath, "--extract-audio", "--audio-format", "mp3")
	} else {
		err = ytdlp.Download(ctx, url, formatID, outPath, "--merge-output-format", "mp4")
	}
	
	if err != nil {
		return "", err
	}
	return outPath, nil
}

func (p *PinterestAdapter) downloadImage(ctx context.Context, pageURL, tempDir string) (string, error) {
	raw, err := ytdlp.Resolve(ctx, pageURL)
	if err != nil {
		return "", err
	}

	imageURL := raw.URL
	if imageURL == "" && len(raw.Formats) > 0 {
		imageURL = raw.Formats[len(raw.Formats)-1].URL
	}
	if imageURL == "" {
		return "", fmt.Errorf("could not find image URL")
	}
	if err := security.ValidateFetchURL(imageURL); err != nil {
		return "", fmt.Errorf("unsafe image URL")
	}

	client := security.SafeHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("image request failed")
	}
	if resp.ContentLength > 50*1024*1024 {
		return "", fmt.Errorf("image is too large")
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return "", fmt.Errorf("resolved URL is not an image")
	}

	ext := "jpg"
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "png") {
		ext = "png"
	} else if strings.Contains(ct, "webp") {
		ext = "webp"
	}

	outPath := filepath.Join(tempDir, fmt.Sprintf("image.%s", ext))
	f, err := createFile(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	written, err := io.Copy(f, io.LimitReader(resp.Body, 50*1024*1024+1))
	if err != nil {
		return "", err
	}
	if written > 50*1024*1024 {
		return "", fmt.Errorf("image is too large")
	}

	return outPath, nil
}
