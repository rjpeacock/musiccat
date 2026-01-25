package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// SearchReleaseGroups searches for release groups by artist ID on MusicBrainz
func SearchReleaseGroups(artistID string) ([]ReleaseGroup, error) {
	// Simple artist search without any filters
	query := fmt.Sprintf("artist:%s", artistID)
	reqURL := fmt.Sprintf("%s/release-group?query=%s&limit=100&fmt=json", baseURL, url.QueryEscape(query))
	
	resp, err := MakeRequest(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ReleaseGroupSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.ReleaseGroups, nil
}

// SearchReleaseGroupsWithFilters searches for release groups with various filters applied
func SearchReleaseGroupsWithFilters(artistID string, albumOnly, singleOnly bool, afterYear, beforeYear int, titleFilter string) ([]ReleaseGroup, error) {
	// Build query with filters
	query := fmt.Sprintf("artist:%s", artistID)

	// Add type filters
	if albumOnly && !singleOnly {
		query += " AND type:album"
	} else if singleOnly && !albumOnly {
		query += " AND type:single"
	}

	// Add year filters
	if afterYear > 0 {
		query += fmt.Sprintf(" AND date:[%d-01-01 TO]", afterYear)
	}
	if beforeYear > 0 {
		if afterYear > 0 {
			query = strings.TrimSuffix(query, " TO]")
			query += fmt.Sprintf(" %d-12-31]", beforeYear-1)
		} else {
			query += fmt.Sprintf(" AND date:[TO %d-12-31]", beforeYear-1)
		}
	}

	// Add title filter
	if titleFilter != "" {
		query += fmt.Sprintf(" AND release:%s", titleFilter)
	}

	// Request up to 100 release groups (MusicBrainz max)
	reqURL := fmt.Sprintf("%s/release-group?query=%s&limit=100&fmt=json", baseURL, url.QueryEscape(query))
	
	resp, err := MakeRequest(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ReleaseGroupSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.ReleaseGroups, nil
}
