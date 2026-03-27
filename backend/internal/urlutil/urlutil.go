package urlutil

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseURL parses and normalizes a URL into a canonical form suitable for persistence and comparison.
func ParseURL(rawURL string) (*url.URL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("URL is required")
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q: only http and https are allowed", u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("URL must have a hostname")
	}
	u.Host = strings.ToLower(u.Host)
	u.Path = strings.TrimRight(u.Path, "/")
	return u, nil
}

// RootURL returns the scheme + host portion of a URL (e.g. "https://example.com").
func RootURL(u *url.URL) string {
	return u.Scheme + "://" + u.Host
}
