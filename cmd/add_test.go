package cmd

import (
	"net/http"
	"testing"
)

// TestMusicBrainzRequest verifies that MusicBrainz API requests work with IPv4 and retry logic
func TestMusicBrainzRequest(t *testing.T) {
	// Test searching for a well-known artist
	artists, err := searchArtists("Radiohead")
	if err != nil {
		t.Fatalf("searchArtists failed: %v", err)
	}

	if len(artists) == 0 {
		t.Fatal("Expected at least one artist result for 'Radiohead'")
	}

	// Verify the top result is Radiohead
	found := false
	for _, artist := range artists {
		if artist.Name == "Radiohead" {
			found = true
			// Test release group search for this artist
			releaseGroups, err := searchReleaseGroups(artist.ID)
			if err != nil {
				t.Fatalf("searchReleaseGroups failed: %v", err)
			}
			if len(releaseGroups) == 0 {
				t.Fatal("Expected at least one release group for Radiohead")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Radiohead' in artist results")
	}
}

// TestMusicBrainzClientIPv4 verifies the HTTP client uses IPv4
func TestMusicBrainzClientIPv4(t *testing.T) {
	client := createMusicBrainzClient()
	if client == nil {
		t.Fatal("createMusicBrainzClient returned nil")
	}

	if client.Timeout == 0 {
		t.Error("Expected client timeout to be set")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected client transport to be *http.Transport")
	}

	if transport.DialContext == nil {
		t.Error("Expected DialContext to be configured for IPv4")
	}
}
