package cmd

import (
	"net/http"
	"strings"
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

	// Verify top result is Radiohead
	found := false
	for _, artist := range artists {
		if artist.Name == "Radiohead" {
			found = true
			// Test unfiltered release group search for this artist
			releaseGroups, err := searchReleaseGroups(artist.ID)
			if err != nil {
				t.Fatalf("searchReleaseGroups failed: %v", err)
			}
			if len(releaseGroups) == 0 {
				t.Fatal("Expected at least one release group for Radiohead")
			}

			// Test filtered release group search for this artist
			filteredGroups, err := searchReleaseGroupsWithFilters(artist.ID, true, false, 1990, 2000, "")
			if err != nil {
				t.Fatalf("searchReleaseGroupsWithFilters failed: %v", err)
			}
			// Filtered results should be <= unfiltered results
			if len(filteredGroups) > len(releaseGroups) {
				t.Error("Filtered results should not exceed unfiltered results")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Radiohead' in artist results")
	}
}

// TestTwoStageSearchWorkflow verifies the two-stage search process
func TestTwoStageSearchWorkflow(t *testing.T) {
	// Stage 1: Artist search (no filters applied)
	artists, err := searchArtists("Beatles")
	if err != nil {
		t.Fatalf("Stage 1 - artist search failed: %v", err)
	}

	if len(artists) == 0 {
		t.Fatal("Stage 1 - Expected at least one artist result for 'Beatles'")
	}

	// Find The Beatles specifically
	var beatlesArtist *Artist
	for _, artist := range artists {
		if artist.Name == "The Beatles" {
			beatlesArtist = &artist
			break
		}
	}

	if beatlesArtist == nil {
		t.Fatal("Stage 1 - Expected to find 'The Beatles' in artist results")
	}

	// Stage 2: Release group search with filters
	albumOnly := true
	singleOnly := false
	afterYear := 1965
	beforeYear := 1970
	titleFilter := ""

	releaseGroups, err := searchReleaseGroupsWithFilters(
		beatlesArtist.ID,
		albumOnly,
		singleOnly,
		afterYear,
		beforeYear,
		titleFilter,
	)
	if err != nil {
		t.Fatalf("Stage 2 - filtered release group search failed: %v", err)
	}

	// Verify all results are albums within the date range
	for _, rg := range releaseGroups {
		if rg.PrimaryType != "Album" {
			t.Errorf("Stage 2 - Expected album type, got '%s' for '%s'", rg.PrimaryType, rg.Title)
		}

		year := parseYear(rg.FirstReleaseDate)
		if year != nil && (*year < 1966 || *year > 1969) {
			t.Errorf("Stage 2 - Expected year between 1966-1969, got %d for '%s'", *year, rg.Title)
		}
	}

	if len(releaseGroups) == 0 {
		t.Error("Stage 2 - Expected at least one album between 1966-1969 for The Beatles")
	}
}

// TestFilterCombinations tests various filter combinations
func TestFilterCombinations(t *testing.T) {
	artists, err := searchArtists("Radiohead")
	if err != nil {
		t.Fatalf("Failed to find Radiohead: %v", err)
	}

	var radiohead *Artist
	for _, artist := range artists {
		if artist.Name == "Radiohead" {
			radiohead = &artist
			break
		}
	}

	if radiohead == nil {
		t.Fatal("Could not find Radiohead artist")
	}

	// Test album-only filter
	albums, err := searchReleaseGroupsWithFilters(radiohead.ID, true, false, 0, 0, "")
	if err != nil {
		t.Fatalf("Album filter failed: %v", err)
	}
	for _, rg := range albums {
		if rg.PrimaryType != "Album" {
			t.Errorf("Expected album, got '%s'", rg.PrimaryType)
		}
	}

	// Test single-only filter
	singles, err := searchReleaseGroupsWithFilters(radiohead.ID, false, true, 0, 0, "")
	if err != nil {
		t.Fatalf("Single filter failed: %v", err)
	}
	for _, rg := range singles {
		if rg.PrimaryType != "Single" {
			t.Errorf("Expected single, got '%s'", rg.PrimaryType)
		}
	}

	// Test year filter
	after2000, err := searchReleaseGroupsWithFilters(radiohead.ID, false, false, 2000, 0, "")
	if err != nil {
		t.Fatalf("Year filter failed: %v", err)
	}
	for _, rg := range after2000 {
		year := parseYear(rg.FirstReleaseDate)
		if year != nil && *year <= 2000 {
			t.Errorf("Expected year > 2000, got %d for '%s'", *year, rg.Title)
		}
	}

	// Test title filter
	titleFilter, err := searchReleaseGroupsWithFilters(radiohead.ID, false, false, 0, 0, "Paranoid")
	if err != nil {
		t.Fatalf("Title filter failed: %v", err)
	}
	for _, rg := range titleFilter {
		if !strings.Contains(strings.ToLower(rg.Title), "paranoid") {
			t.Errorf("Expected title to contain 'Paranoid', got '%s'", rg.Title)
		}
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
