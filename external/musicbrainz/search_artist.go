package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const baseURL = "https://musicbrainz.org/ws/2"

// SearchArtists searches for artists by name on MusicBrainz
func SearchArtists(query string) ([]Artist, error) {
	encodedQuery := url.QueryEscape(query)
	reqURL := fmt.Sprintf("%s/artist?query=artist:%s&fmt=json", baseURL, encodedQuery)
	
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
