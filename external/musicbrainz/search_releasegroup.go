package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// SearchReleaseGroups searches for release groups by artist ID on MusicBrainz.
// All filter parameters are optional - pass zero values to skip filtering.
func SearchReleaseGroups(artistID string, albumOnly, singleOnly bool, afterYear, beforeYear int, titleFilter string) ([]ReleaseGroup, error) {
	// Build query with filters
	query := fmt.Sprintf("arid:%s", artistID)

	// Add type filters
	if albumOnly && !singleOnly {
		query += " AND type:album"
	} else if singleOnly && !albumOnly {
		query += " AND type:single"
	}

	// Add year filters
	if afterYear > 0 && beforeYear > 0 {
		query += fmt.Sprintf(" AND date:[%d-01-01 TO %d-12-31]", afterYear, beforeYear-1)
	} else if afterYear > 0 {
		query += fmt.Sprintf(" AND date:[%d-01-01 TO]", afterYear)
	} else if beforeYear > 0 {
		query += fmt.Sprintf(" AND date:[TO %d-12-31]", beforeYear-1)
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
