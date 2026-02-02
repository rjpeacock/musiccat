package cmd

import (
	"net/http"
	"strings"
	"testing"

	"musiccat/cmd/helpers"
	"musiccat/external/musicbrainz"
)

// TestMusicBrainzRequest verifies that MusicBrainz API requests work with IPv4 and retry logic
func TestMusicBrainzRequest(t *testing.T) {
	// Test searching for a well-known artist
	artists, err := musicbrainz.SearchArtists("Radiohead", 25, 0)
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
			releaseGroups, err := musicbrainz.SearchReleaseGroups(artist.ID, false, false, false, false, false, false, 0, "", 0)
			if err != nil {
				t.Fatalf("searchReleaseGroups failed: %v", err)
			}
			if len(releaseGroups) == 0 {
				t.Fatal("Expected at least one release group for Radiohead")
			}

			// Test filtered release group search for this artist
			filteredGroups, err := musicbrainz.SearchReleaseGroups(artist.ID, true, false, false, false, false, false, 1995, "", 0)
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
	artists, err := musicbrainz.SearchArtists("Beatles", 25, 0)
	if err != nil {
		t.Fatalf("Stage 1 - artist search failed: %v", err)
	}

	if len(artists) == 0 {
		t.Fatal("Stage 1 - Expected at least one artist result for 'Beatles'")
	}

	// Find The Beatles specifically
	var beatlesArtist *musicbrainz.Artist
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
	epOnly := false
	compilationOnly := false
	liveOnly := false
	soundtrackOnly := false
	year := 1967
	titleFilter := ""

	releaseGroups, err := musicbrainz.SearchReleaseGroups(
		beatlesArtist.ID,
		albumOnly,
		singleOnly,
		epOnly,
		compilationOnly,
		liveOnly,
		soundtrackOnly,
		year,
		titleFilter,
		0,
	)
	if err != nil {
		t.Fatalf("Stage 2 - filtered release group search failed: %v", err)
	}

	// Verify all results are albums from 1967
	for _, rg := range releaseGroups {
		if rg.PrimaryType != "Album" {
			t.Errorf("Stage 2 - Expected album type, got '%s' for '%s'", rg.PrimaryType, rg.Title)
		}

		releaseYear := helpers.ParseYear(rg.FirstReleaseDate)
		if releaseYear != nil && *releaseYear != 1967 {
			t.Errorf("Stage 2 - Expected year 1967, got %d for '%s'", *releaseYear, rg.Title)
		}
	}

	if len(releaseGroups) == 0 {
		t.Error("Stage 2 - Expected at least one album from 1967 for The Beatles")
	}
}

// TestFilterCombinations tests various filter combinations
func TestFilterCombinations(t *testing.T) {
	artists, err := musicbrainz.SearchArtists("Radiohead", 25, 0)
	if err != nil {
		t.Fatalf("Failed to find Radiohead: %v", err)
	}

	var radiohead *musicbrainz.Artist
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
	albums, err := musicbrainz.SearchReleaseGroups(radiohead.ID, true, false, false, false, false, false, 0, "", 0)
	if err != nil {
		t.Fatalf("Album filter failed: %v", err)
	}
	for _, rg := range albums {
		if rg.PrimaryType != "Album" {
			t.Errorf("Expected album, got '%s'", rg.PrimaryType)
		}
	}

	// Test single-only filter
	singles, err := musicbrainz.SearchReleaseGroups(radiohead.ID, false, true, false, false, false, false, 0, "", 0)
	if err != nil {
		t.Fatalf("Single filter failed: %v", err)
	}
	for _, rg := range singles {
		if rg.PrimaryType != "Single" {
			t.Errorf("Expected single, got '%s'", rg.PrimaryType)
		}
	}

	// Test year filter
	year2007, err := musicbrainz.SearchReleaseGroups(radiohead.ID, false, false, false, false, false, false, 2007, "", 0)
	if err != nil {
		t.Fatalf("Year filter failed: %v", err)
	}
	for _, rg := range year2007 {
		releaseYear := helpers.ParseYear(rg.FirstReleaseDate)
		if releaseYear != nil && *releaseYear != 2007 {
			t.Errorf("Expected year 2007, got %d for '%s'", *releaseYear, rg.Title)
		}
	}

	// Test title filter
	titleFilter, err := musicbrainz.SearchReleaseGroups(radiohead.ID, false, false, false, false, false, false, 0, "Paranoid", 0)
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
	client := musicbrainz.CreateClient()
	if client == nil {
		t.Fatal("CreateClient returned nil")
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
