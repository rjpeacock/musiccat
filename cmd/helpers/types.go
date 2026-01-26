package helpers

import "musiccat/external/musicbrainz"

type SelectionItem struct {
	ReleaseGroup musicbrainz.ReleaseGroup
	Promo        bool
	Pirate       bool
	Quantity     int
	Notes        string
}
