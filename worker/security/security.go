package security

import (
	"context"
	cryptoRand "crypto/rand"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var safeFilenamePattern = regexp.MustCompile(`[^\pL\pN _.-]+`)
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = cryptoRand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

var allowedHosts = map[string]string{
	"youtube.com": "youtube", "www.youtube.com": "youtube", "m.youtube.com": "youtube", "youtu.be": "youtube",
	"pinterest.com": "pinterest", "www.pinterest.com": "pinterest", "pin.it": "pinterest",
	"instagram.com": "instagram", "www.instagram.com": "instagram",
	"tiktok.com": "tiktok", "www.tiktok.com": "tiktok",
	"x.com": "x", "www.x.com": "x", "twitter.com": "x", "www.twitter.com": "x",
}

func AuthenticateBearer(r *http.Request, secret string) bool {
	if secret == "" {
		return false
	}
	provided := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if provided == "" || len(provided) != len(secret) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) == 1
}

func ValidateMediaURL(raw string) (string, error) {
	parsed, err := parseURL(raw)
	if err != nil {
		return "", err
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	platform, ok := allowedHosts[host]
	if !ok {
		return "", fmt.Errorf("unsupported media host")
	}
	if err := validatePublicHost(host); err != nil {
		return "", err
	}
	return platform, nil
}

func PlatformForURL(raw string) (string, error) {
	parsed, err := parseURL(raw)
	if err != nil {
		return "", err
	}
	platform, ok := allowedHosts[strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))]
	if !ok {
		return "", fmt.Errorf("unsupported media host")
	}
	return platform, nil
}

func ValidateFetchURL(raw string) error {
	parsed, err := parseURL(raw)
	if err != nil {
		return err
	}
	return validatePublicHost(strings.ToLower(strings.TrimSuffix(parsed.Hostname(), ".")))
}

func SafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: timeout}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			ips, err := publicIPs(host)
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if dialErr == nil {
					return conn, nil
				}
			}
			return nil, fmt.Errorf("could not connect to external host")
		},
		TLSHandshakeTimeout: timeout,
		ResponseHeaderTimeout: timeout,
		IdleConnTimeout: 30 * time.Second,
	}
	return &http.Client{
		Timeout: timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			return ValidateFetchURL(req.URL.String())
		},
	}
}

func ValidateUUID(value string) error {
	if !uuidPattern.MatchString(value) {
		return fmt.Errorf("invalid job ID")
	}
	return nil
}

func ValidateOutput(output string) bool {
	switch output {
	case "mp4", "mp3", "jpg", "png", "webp":
		return true
	default:
		return false
	}
}

func ValidateFormatID(value string) bool {
	if value == "original" {
		return true
	}
	if value == "" || len(value) > 128 || strings.ContainsAny(value, `/\\\x00\r\n`) {
		return false
	}
	for _, char := range value {
		if !(char == '+' || char == '-' || char == '_' || char == '.' || char >= '0' && char <= '9' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z') {
			return false
		}
	}
	return true
}

func SafeFilename(title, extension string) string {
	base := safeFilenamePattern.ReplaceAllString(title, "")
	base = strings.TrimSpace(strings.Trim(base, "."))
	base = strings.ReplaceAll(base, "..", "")
	if base == "" {
		base = "media"
	}
	extension = strings.TrimPrefix(strings.ToLower(extension), ".")
	if !ValidateOutput(extension) {
		extension = "mp4"
	}
	return filepath.Base(base + "." + extension)
}

func parseURL(raw string) (*url.URL, error) {
	if len(raw) == 0 || len(raw) > 2048 {
		return nil, fmt.Errorf("invalid URL")
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, fmt.Errorf("invalid URL")
	}
	if parsed.Port() != "" && parsed.Port() != "80" && parsed.Port() != "443" {
		return nil, fmt.Errorf("unsupported URL port")
	}
	return parsed, nil
}

func validatePublicHost(host string) error {
	if host == "" || strings.EqualFold(host, "localhost") {
		return fmt.Errorf("internal host is not allowed")
	}
	_, err := publicIPs(host)
	return err
}

func publicIPs(host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("internal address is not allowed")
		}
		return []net.IP{ip}, nil
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("host could not be resolved")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("host resolves to an internal address")
		}
	}
	return ips, nil
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}
