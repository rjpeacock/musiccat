package musicbrainz

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// GetReleaseGroup fetches a specific release group by ID from MusicBrainz
func GetReleaseGroup(mbid string) (*ReleaseGroup, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group/%s?fmt=json", mbid)

	// Rate limit: 1 request per second
	time.Sleep(1 * time.Second)

	resp, err := MakeRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("MusicBrainz API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var releaseGroup ReleaseGroup
	if err := json.Unmarshal(body, &releaseGroup); err != nil {
		return nil, err
	}

	return &releaseGroup, nil
}
