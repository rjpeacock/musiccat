package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// SearchReleaseGroups searches for release groups by artist ID on MusicBrainz.
// All filter parameters are optional - pass zero values to skip filtering.
// offset specifies the starting position for pagination (0 for first page, 100 for second, etc.)
func SearchReleaseGroups(artistID string, albumOnly, singleOnly bool, year int, titleFilter string, offset int) ([]ReleaseGroup, error) {
	// Build query with filters
	query := fmt.Sprintf("arid:%s", artistID)

	// Add type filters
	if albumOnly && !singleOnly {
		query += " AND type:album"
	} else if singleOnly && !albumOnly {
		query += " AND type:single"
	}

	// Add year filter
	if year > 0 {
		query += fmt.Sprintf(" AND firstreleasedate:%d", year)
	}

	// Add title filter
	if titleFilter != "" {
		query += fmt.Sprintf(" AND release:%s", titleFilter)
	}

	// Request up to 100 release groups (MusicBrainz max)
	reqURL := fmt.Sprintf("%s/release-group?query=%s&limit=100&offset=%d&fmt=json", baseURL, url.QueryEscape(query), offset)
	
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
