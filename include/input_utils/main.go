package input_utils

import (
	"net"
	"net/url"
	"os"
)

// IsFile checks if path exists and is a regular file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDomainOrURL checks if the input looks like a domain or URL
func IsDomainOrURL(input string) bool {
	// Try URL parse
	if u, err := url.Parse(input); err == nil {
		// Case 1: Full URL with scheme (http://example.com)
		if u.Scheme != "" && u.Host != "" {
			return true
		}
	}

	// Case 2: Plain domain name (example.com)
	if _, err := net.LookupHost(input); err == nil {
		return true
	}

	return false
}
