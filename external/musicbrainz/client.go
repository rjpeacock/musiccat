package musicbrainz

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// CreateClient creates an HTTP client configured for IPv4-only connections to MusicBrainz
func CreateClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Force IPv4 by replacing "tcp" with "tcp4"
			return dialer.DialContext(ctx, "tcp4", addr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// MakeRequest performs an HTTP GET request with retry logic and rate limiting
func MakeRequest(url string) (*http.Response, error) {
	const maxRetries = 3
	var lastErr error

	client := CreateClient()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "musiccat/0.1 (robertjamespeacock@gmail.com)")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				// Exponential backoff: 1s, 2s
				waitTime := time.Duration(attempt) * time.Second
				fmt.Printf("Request failed (attempt %d/%d), retrying in %v...\n", attempt, maxRetries, waitTime)
				time.Sleep(waitTime)
				continue
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("API error: %s", resp.Status)
			if attempt < maxRetries {
				waitTime := time.Duration(attempt) * time.Second
				fmt.Printf("API returned %s (attempt %d/%d), retrying in %v...\n", resp.Status, attempt, maxRetries, waitTime)
				time.Sleep(waitTime)
				continue
			}
			return nil, lastErr
		}

		// Success - apply rate limiting before returning
		time.Sleep(1 * time.Second)
		return resp, nil
	}

	return nil, lastErr
}
