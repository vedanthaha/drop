package platforms

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
	for name, adapter := range adapters {
		if adapter.CanHandle(url) {
			return name
		}
	}
	return "unknown"
}
