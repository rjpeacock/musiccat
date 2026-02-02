package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const baseURL = "https://musicbrainz.org/ws/2"

// SearchArtists searches for artists by name on MusicBrainz
// limit: number of results to fetch (0 = default 25, max 100)
// offset: starting position for pagination
func SearchArtists(query string, limit int, offset int) ([]Artist, error) {
	encodedQuery := url.QueryEscape(query)
	
	// Default limit is 25, max is 100
	if limit == 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	
	reqURL := fmt.Sprintf("%s/artist?query=artist:%s&limit=%d&offset=%d&fmt=json", baseURL, encodedQuery, limit, offset)
	
	resp, err := MakeRequest(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ArtistSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Artists, nil
}
