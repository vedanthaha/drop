package platforms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
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
		// Fallback for image pins (yt-dlp fails on non-video pins)
		imgURL, imgErr := p.scrapePinterestImage(ctx, url)
		if imgErr != nil {
			return nil, fmt.Errorf("could not resolve Pinterest content: %v (fallback failed: %v)", err, imgErr)
		}
		
		return p.buildImageResult(security.GenerateUUID(), url, imgURL, "Pinterest Image", 1000, 1000), nil
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

	// Image pin (yt-dlp managed to parse something without error, but no video format)
	imageURL := raw.URL
	if imageURL == "" && len(raw.Formats) > 0 {
		imageURL = raw.Formats[len(raw.Formats)-1].URL
	}
	return p.buildImageResult(raw.ID, url, imageURL, raw.Title, raw.Width, raw.Height), nil
}

func (p *PinterestAdapter) buildImageResult(id, sourceURL, imageURL, title string, width, height int) *MediaResult {
	if width == 0 { width = 1000 }
	if height == 0 { height = 1000 }
	
	formats := []MediaFormat{
		{
			ID:    "original",
			Label: fmt.Sprintf("Original %d\u00d7%d", width, height),
			Width: width, Height: height,
			Ext: "jpg",
		},
		{
			ID:    "upscale-2x",
			Label: fmt.Sprintf("Upscale 2x (%d\u00d7%d)", width*2, height*2),
			Width: width * 2, Height: height * 2,
			Ext: "jpg",
		},
		{
			ID:    "upscale-4x",
			Label: fmt.Sprintf("Upscale 4x (%d\u00d7%d)", width*4, height*4),
			Width: width * 4, Height: height * 4,
			Ext: "jpg",
		},
	}

	return &MediaResult{
		ID:        id,
		Platform:  "pinterest",
		Type:      "image",
		Title:     title,
		Thumbnail: imageURL,
		Width:     width,
		Height:    height,
		Formats:   formats,
	}
}

func (p *PinterestAdapter) scrapePinterestImage(ctx context.Context, pageURL string) (string, error) {
	client := security.SafeHTTPClient(10 * time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024*2)) // read up to 2MB
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)
	
	// Try to find og:image regardless of attribute order
	// <meta content="https://..." ... property="og:image"/>
	propTag := `property="og:image"`
	idx := strings.Index(body, propTag)
	if idx >= 0 {
		// Look backwards for <meta
		startMeta := strings.LastIndex(body[:idx], "<meta")
		if startMeta >= 0 {
			// Find content="..." inside this tag
			endMeta := strings.Index(body[idx:], ">")
			if endMeta >= 0 {
				tag := body[startMeta : idx+endMeta]
				contentIdx := strings.Index(tag, `content="`)
				if contentIdx >= 0 {
					startContent := contentIdx + 9
					endContent := strings.Index(tag[startContent:], `"`)
					if endContent > 0 {
						imgUrl := tag[startContent : startContent+endContent]
						// Pinterest thumbnails are usually 736x or originals
						imgUrl = strings.ReplaceAll(imgUrl, "736x", "originals")
						return imgUrl, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("og:image not found")
}

func (p *PinterestAdapter) Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error) {
	if strings.HasPrefix(formatID, "upscale-") || formatID == "original" {
		origPath, err := p.downloadImage(ctx, url, tempDir)
		if err != nil {
			return "", err
		}
		if formatID == "original" {
			return origPath, nil
		}

		scale := 2
		if formatID == "upscale-4x" {
			scale = 4
		}

		ext := filepath.Ext(origPath)
		outPath := filepath.Join(tempDir, fmt.Sprintf("upscaled%s", ext))

		cmd := exec.CommandContext(ctx, "ffmpeg", "-i", origPath, "-vf", fmt.Sprintf("scale=iw*%d:ih*%d", scale, scale), outPath)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("upscale failed: %w", err)
		}
		return outPath, nil
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
	imageURL := ""
	raw, err := ytdlp.Resolve(ctx, pageURL)
	if err == nil {
		imageURL = raw.URL
		if imageURL == "" && len(raw.Formats) > 0 {
			imageURL = raw.Formats[len(raw.Formats)-1].URL
		}
	} else {
		fallbackURL, fErr := p.scrapePinterestImage(ctx, pageURL)
		if fErr != nil {
			return "", fmt.Errorf("yt-dlp resolve failed: %v (fallback failed: %v)", err, fErr)
		}
		imageURL = fallbackURL
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
