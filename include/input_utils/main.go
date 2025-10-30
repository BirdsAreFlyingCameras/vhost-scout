package input_utils

import (
	"net"
	"net/url"
	"os"
)

// IsFile checks if path exists and is a regular file
func IsFile(path string) bool {
	info, file_stat_err := os.Stat(path)
	if file_stat_err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDomainOrURL checks if the input looks like a domain or URL
func IsDomainOrURL(input string) bool {
	// Try URL parse
	parsed_url, url_parsing_err := url.Parse(input)
	if url_parsing_err == nil {
		// Case 1: Full URL with scheme (http://example.com)
		if parsed_url.Scheme != "" && parsed_url.Host != "" {
			return true
		}
	}

	// Case 2: Plain domain name (example.com)
	_, host_name_lookup_err := net.LookupHost(input)
	if host_name_lookup_err == nil {
		return true
	}

	return false
}
