package request_utils

import (
	"math/rand"
	"net/http"
)

// weightedRandom selects a random item based on probabilities.
func weighted_random(weights map[string]float64) string {
	// Compute total weight
	total := 0.0
	for _, w := range weights {
		total += w
	}

	// Pick a random number in [0, total)
	r := rand.Float64() * total

	// Iterate through items until we reach the random threshold
	cumulative := 0.0
	for item, weight := range weights {
		cumulative += weight
		if r < cumulative {
			return item
		}
	}

	// Fallback (shouldn't happen if weights are valid)
	for item := range weights {
		return item
	}
	return ""
}

func Generate_random_request_headers() http.Header {

	acceptHeaders := []string{
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"application/json",
		"*/*",
		"image/avif,image/webp,image/apng,image/*;q=0.8,*/*;q=0.5",
		"application/xml;q=0.9,text/xml;q=0.8,*/*;q=0.7",
		"text/plain;q=0.9,application/json;q=0.8,*/*;q=0.5",
	}

	userAgents := map[string]float64{ // All win10 latest user agents
		"Mozilla / 5.0(Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36":                  0.79, // Chrome (79% market share)
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.3537.92": 0.11, // Edge (11.5% market share)
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:144.0) Gecko/20100101 Firefox/144.0":                                                  0.04, // FireFox (4% market share)
	}

	languages := []string{
		"en-US,en;q=0.9",
		"es-ES,en;q=0.8",
		"fr-FR,en;q=0.7",
		"de-DE,en;q=0.7",
		"ja-JP,en;q=0.6",
	}

	headers := http.Header{}
	headers.Set("Accept", acceptHeaders[rand.Intn(len(acceptHeaders))])
	headers.Set("Accept-Language", languages[rand.Intn(len(languages))])
	headers.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	headers.Set("Cache-Control", "max-age=0")
	headers.Set("User-Agent", weighted_random(userAgents))
	return headers
}
