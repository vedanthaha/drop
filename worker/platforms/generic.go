package platforms

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/alldownload/worker/security"
	"github.com/alldownload/worker/ytdlp"
)

// GenericAdapter handles Instagram, TikTok, and X using yt-dlp
type GenericAdapter struct {
	PlatformName string
	URLPatterns  []string
}

func (g *GenericAdapter) CanHandle(url string) bool {
	platform, err := security.PlatformForURL(url)
	return err == nil && platform == g.PlatformName
}

func (g *GenericAdapter) Resolve(ctx context.Context, url string) (*MediaResult, error) {
	raw, err := ytdlp.Resolve(ctx, url)
	if err != nil {
		return nil, err
	}

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
		if f.VCodec == "none" || f.VCodec == "" || f.Height == 0 || f.Height > 1080 {
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

	mediaType := "video"

	return &MediaResult{
		ID:        raw.ID,
		Platform:  g.PlatformName,
		Type:      mediaType,
		Title:     raw.Title,
		Thumbnail: raw.Thumbnail,
		Width:     raw.Width,
		Height:    raw.Height,
		Duration:  int(raw.Duration),
		Formats:   formats,
	}, nil
}

func (g *GenericAdapter) Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error) {
	outPath := filepath.Join(tempDir, fmt.Sprintf("media.%s", output))
	var err error
	if output == "mp3" {
		err = ytdlp.Download(ctx, url, "bestaudio/best", outPath, "--extract-audio", "--audio-format", "mp3")
	} else {
		formatSpec := fmt.Sprintf("%s+bestaudio/best", formatID)
		err = ytdlp.Download(ctx, url, formatSpec, outPath, "--merge-output-format", "mp4")
	}

	if err != nil {
		return "", err
	}
	return outPath, nil
}
