package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"musiccat/external/musicbrainz"
)

// SelectArtist displays artists with pagination and returns the selected artist
func SelectArtist(artistName string) (*musicbrainz.Artist, error) {
	const pageSize = 25
	offset := 0
	
	for {
		// Fetch artists
		artists, err := musicbrainz.SearchArtists(artistName, pageSize, offset)
		if err != nil {
			return nil, err
		}
		
		if len(artists) == 0 {
			if offset == 0 {
				return nil, fmt.Errorf("no artists found for '%s'", artistName)
			}
			fmt.Println("No more artists found.")
			offset -= pageSize
			if offset < 0 {
				offset = 0
			}
			continue
		}
		
		// Display artists
		fmt.Printf("\nArtists found (showing %d-%d):\n", offset+1, offset+len(artists))
		for i, artist := range artists {
			disamb := ""
			if artist.Disambiguation != "" {
				disamb = " (" + artist.Disambiguation + ")"
			}
			fmt.Printf("%2d. %s%s\n", i+1, artist.Name, disamb)
		}
		
		// Prompt for selection
		fmt.Print("\nSelect artist (number), 00 for more, or 99 to go back: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return nil, scanner.Err()
		}
		
		input := strings.TrimSpace(scanner.Text())
		
		// Handle special commands
		if input == "00" {
			offset += pageSize
			continue
		}
		
		if input == "99" {
			if offset == 0 {
				return nil, fmt.Errorf("cancelled")
			}
			offset -= pageSize
			if offset < 0 {
				offset = 0
			}
			continue
		}
		
		// Parse selection
		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(artists) {
			fmt.Printf("Invalid selection. Enter a number between 1 and %d, 00 for more, or 99 to go back.\n", len(artists))
			continue
		}
		
		return &artists[num-1], nil
	}
}
