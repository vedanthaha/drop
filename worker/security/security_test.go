package security

import (
	"testing"
)

func TestValidateMediaURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
		err      bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "youtube", false},
		{"https://youtu.be/dQw4w9WgXcQ", "youtube", false},
		{"https://www.pinterest.com/pin/123456789/", "pinterest", false},
		{"https://pin.it/123456789", "pinterest", false},
		{"https://www.instagram.com/p/123456789/", "instagram", false},
		{"https://www.tiktok.com/@user/video/123456789", "tiktok", false},
		{"https://x.com/status/123456789", "x", false},
		{"https://twitter.com/status/123456789", "x", false},
		{"http://localhost:3000", "", true},
		{"http://127.0.0.1", "", true},
		{"http://192.168.1.1", "", true},
		{"https://example.com/?youtube.com", "", true},
		{"ftp://youtube.com", "", true},
		{"invalid-url", "", true},
	}

	for _, tt := range tests {
		platform, err := ValidateMediaURL(tt.url)
		if tt.err {
			if err == nil {
				t.Errorf("expected error for url %s, got none", tt.url)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for url %s: %v", tt.url, err)
			}
			if platform != tt.expected {
				t.Errorf("expected platform %s for url %s, got %s", tt.expected, tt.url, platform)
			}
		}
	}
}

func TestValidateFormatID(t *testing.T) {
	tests := []struct {
		formatID string
		expected bool
	}{
		{"original", true},
		{"137+140", true},
		{"137", true},
		{"bestvideo+bestaudio/best", false},
		{"", false},
		{"foo;bar", false},
		{"foo bar", false},
		{"foo\nbar", false},
		{"foo\"bar\"", false},
		{"foo|bar", false},
		{"foo<bar>", false},
	}

	for _, tt := range tests {
		valid := ValidateFormatID(tt.formatID)
		if valid != tt.expected {
			t.Errorf("expected %v for format ID %s, got %v", tt.expected, tt.formatID, valid)
		}
	}
}

func TestSafeFilename(t *testing.T) {
	tests := []struct {
		title     string
		extension string
		expected  string
	}{
		{"hello world", "mp4", "hello world.mp4"},
		{"hello/world", "mp3", "helloworld.mp3"},
		{"hello\\world", "jpg", "helloworld.jpg"},
		{"hello:world", "png", "helloworld.png"},
		{"hello*world", "webp", "helloworld.webp"},
		{"hello?world", "mp4", "helloworld.mp4"},
		{"hello\"world\"", "mp4", "helloworld.mp4"},
		{"hello<world>", "mp4", "helloworld.mp4"},
		{"hello|world", "mp4", "helloworld.mp4"},
		{"hello..world", "mp4", "helloworld.mp4"},
		{"", "mp4", "media.mp4"},
		{"hello world", "invalid", "hello world.mp4"},
	}

	for _, tt := range tests {
		filename := SafeFilename(tt.title, tt.extension)
		if filename != tt.expected {
			t.Errorf("expected %s for title %s and extension %s, got %s", tt.expected, tt.title, tt.extension, filename)
		}
	}
}
