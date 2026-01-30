package musicbrainz

// ArtistSearchResponse represents the response from MusicBrainz artist search
type ArtistSearchResponse struct {
	Artists []Artist `json:"artists"`
}

// Artist represents a MusicBrainz artist
type Artist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
}

// ReleaseGroupSearchResponse represents the response from MusicBrainz release group search
type ReleaseGroupSearchResponse struct {
	ReleaseGroups []ReleaseGroup `json:"release-groups"`
}

// ReleaseGroup represents a MusicBrainz release group
type ReleaseGroup struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	FirstReleaseDate string   `json:"first-release-date"`
	PrimaryType      string   `json:"primary-type"`
	SecondaryTypes   []string `json:"secondary-types"`
}
