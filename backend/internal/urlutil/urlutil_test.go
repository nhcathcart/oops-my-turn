//go:build unit

package urlutil_test

import (
	"testing"

	"github.com/nhcathcart/oops-my-turn/backend/internal/urlutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURL_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantURL  string
		wantHost string
	}{
		{
			name:     "adds https to bare domain",
			input:    "example.com",
			wantURL:  "https://example.com",
			wantHost: "example.com",
		},
		{
			name:     "lowercases host and strips trailing slash",
			input:    "https://EXAMPLE.COM/about/",
			wantURL:  "https://example.com/about",
			wantHost: "example.com",
		},
		{
			name:     "preserves explicit http scheme",
			input:    "http://example.com/docs/",
			wantURL:  "http://example.com/docs",
			wantHost: "example.com",
		},
		{
			name:     "trims surrounding whitespace",
			input:    "  https://example.com/pricing/  ",
			wantURL:  "https://example.com/pricing",
			wantHost: "example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsedURL, err := urlutil.ParseURL(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.wantURL, parsedURL.String())
			assert.Equal(t, tc.wantHost, parsedURL.Hostname())
		})
	}
}

func TestParseURL_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty",
			input:   "",
			wantErr: "URL is required",
		},
		{
			name:    "unsupported scheme",
			input:   "ftp://example.com",
			wantErr: `unsupported scheme "ftp"`,
		},
		{
			name:    "missing host",
			input:   "https://",
			wantErr: "URL must have a hostname",
		},
		{
			name:    "javascript url",
			input:   "javascript:alert(1)",
			wantErr: "invalid URL:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := urlutil.ParseURL(tc.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestRootURL(t *testing.T) {
	parsedURL, err := urlutil.ParseURL("https://example.com/some/path?q=1")
	require.NoError(t, err)

	assert.Equal(t, "https://example.com", urlutil.RootURL(parsedURL))
}
