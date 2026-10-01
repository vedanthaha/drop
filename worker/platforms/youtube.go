package platforms

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/alldownload/worker/security"
	"github.com/alldownload/worker/ytdlp"
)

type YouTubeAdapter struct{}

func (y *YouTubeAdapter) CanHandle(url string) bool {
	platform, err := security.PlatformForURL(url)
	return err == nil && platform == "youtube"
}

func (y *YouTubeAdapter) Resolve(ctx context.Context, url string) (*MediaResult, error) {
	raw, err := ytdlp.Resolve(ctx, url)
	if err != nil {
		return nil, err
	}

	type formatEntry struct {
		formatID string
		height   int
		width    int
		fps      int
		ext      string
		hasAudio bool
		filesize int64
		tbr      float64
	}

	bestByHeight := make(map[int]*formatEntry)

	for _, f := range raw.Formats {
		if f.VCodec == "none" || f.VCodec == "" || f.Height == 0 {
			continue
		}

		hasAudio := f.ACodec != "none" && f.ACodec != ""
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
				fps:      int(f.FPS),
				ext:      f.Ext,
				hasAudio: hasAudio,
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
			FPS:      e.fps,
			Ext:      e.ext,
			HasAudio: e.hasAudio,
			Filesize: e.filesize,
		})
	}

	return &MediaResult{
		ID:        raw.ID,
		Platform:  "youtube",
		Type:      "video",
		Title:     raw.Title,
		Thumbnail: raw.Thumbnail,
		Width:     raw.Width,
		Height:    raw.Height,
		Duration:  int(raw.Duration),
		Formats:   formats,
	}, nil
}

func (y *YouTubeAdapter) Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error) {
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
