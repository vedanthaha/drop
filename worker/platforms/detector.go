package platforms

import "github.com/alldownload/worker/security"

var adapters = map[string]Adapter{
	"youtube":   &YouTubeAdapter{},
	"pinterest": &PinterestAdapter{},
	"instagram": &GenericAdapter{PlatformName: "instagram", URLPatterns: []string{"instagram.com"}},
	"tiktok":    &GenericAdapter{PlatformName: "tiktok", URLPatterns: []string{"tiktok.com"}},
	"x":         &GenericAdapter{PlatformName: "x", URLPatterns: []string{"x.com", "twitter.com"}},
}

func GetAdapter(platform string) Adapter {
	return adapters[platform]
}

func DetectPlatform(url string) string {
	platform, err := security.ValidateMediaURL(url)
	if err != nil {
		return "unknown"
	}
	if _, ok := adapters[platform]; !ok {
		return "unknown"
	}
	return platform
}
