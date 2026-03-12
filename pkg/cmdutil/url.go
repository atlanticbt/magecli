package cmdutil

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("host is required")
	}
	if strings.HasPrefix(raw, "http://") {
		return "", fmt.Errorf("insecure HTTP is not allowed; use https:// to protect authentication tokens")
	}
	if !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/"), nil
}

func HostKeyFromURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid base URL %q", baseURL)
	}
	return u.Host, nil
}
