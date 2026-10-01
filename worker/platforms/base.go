package platforms

import "context"

type MediaFormat struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FPS      int    `json:"fps"`
	Ext      string `json:"ext"`
	HasAudio bool   `json:"hasAudio"`
	Filesize int64  `json:"filesize"`
}

type MediaResult struct {
	ID        string        `json:"id"`
	Platform  string        `json:"platform"`
	Type      string        `json:"type"`
	Title     string        `json:"title"`
	Thumbnail string        `json:"thumbnail"`
	SourceURL string        `json:"sourceUrl"`
	Width     int           `json:"width"`
	Height    int           `json:"height"`
	Duration  int           `json:"duration"`
	Formats   []MediaFormat `json:"formats"`
}

type Adapter interface {
	CanHandle(url string) bool
	Resolve(ctx context.Context, url string) (*MediaResult, error)
	Download(ctx context.Context, jobID, url, formatID, output, tempDir string) (string, error)
}

// HeightToLabel maps video height to a human-readable quality label
func HeightToLabel(height int) string {
	switch {
	case height >= 2160:
		return "4K"
	case height >= 1440:
		return "1440p"
	case height >= 1080:
		return "1080p"
	case height >= 720:
		return "720p"
	case height >= 480:
		return "480p"
	case height >= 360:
		return "360p"
	case height >= 240:
		return "240p"
	default:
		return "SD"
	}
}
