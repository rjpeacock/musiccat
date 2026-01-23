package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"musiccat/internal/config"
	"musiccat/internal/db"
)

type ArtistSearchResponse struct {
	Artists []Artist `json:"artists"`
}

type Artist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
}

type ReleaseGroupSearchResponse struct {
	ReleaseGroups []ReleaseGroup `json:"release-groups"`
}

type ReleaseGroup struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	FirstReleaseDate string `json:"first-release-date"`
	PrimaryType      string `json:"primary-type"`
}

var addCmd = &cobra.Command{
	Use:   `add "<artist name>"`,
	Short: "Search and add releases by artist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manual, _ := cmd.Flags().GetBool("manual")
		if manual {
			if len(args) != 0 {
				return fmt.Errorf("--manual takes no arguments")
			}
			return addManual()
		}
		if len(args) != 1 {
			return fmt.Errorf("requires artist name argument")
		}
		artistName := args[0]
		pageSize, _ := cmd.Flags().GetInt("page-size")
		return addFromMusicBrainz(artistName, pageSize)
	},
}

func addFromMusicBrainz(artistName string, pageSize int) error {
	// Load config for current format
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.CurrentFormat == "" {
		return fmt.Errorf("no current format set; use 'musiccat set-format <FORMAT>' first")
	}

	// Open DB
	database, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer database.Close()

	// Bootstrap if needed
	if err := db.BootstrapDB(database); err != nil {
		return err
	}

	// Search artists
	artists, err := searchArtists(artistName)
	if err != nil {
		return err
	}
	if len(artists) == 0 {
		return fmt.Errorf("no artists found for '%s'", artistName)
	}

	// Display and select artist
	fmt.Println("Artists found:")
	for i, artist := range artists {
		disamb := ""
		if artist.Disambiguation != "" {
			disamb = " (" + artist.Disambiguation + ")"
		}
		fmt.Printf("%d. %s%s\n", i+1, artist.Name, disamb)
	}
	selectedArtist, err := selectItem("Select artist (number): ", artists)
	if err != nil {
		return err
	}

	// Search release groups
	releaseGroups, err := searchReleaseGroups(selectedArtist.ID)
	if err != nil {
		return err
	}
	if len(releaseGroups) == 0 {
		return fmt.Errorf("no releases found for artist '%s'", selectedArtist.Name)
	}

	selectedItems, err := selectReleasesWithPagination(releaseGroups, pageSize)
	if err != nil {
		return err
	}

	// Insert each selected release and ownership
	for _, item := range selectedItems {
		rg := releaseGroups[item.index-1]
		var releaseYear *int = parseYear(rg.FirstReleaseDate)
		releaseID, err := db.InsertRelease(database, selectedArtist.Name, rg.Title, releaseYear, &rg.ID)
		if err != nil {
			return err
		}
		_, err = db.InsertOwnership(database, releaseID, cfg.CurrentFormat, nil, nil, nil, nil, nil, nil, item.promo)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Added %d releases to collection.\n", len(selectedItems))
	return nil
}

func addManual() error {
	// Open DB
	database, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer database.Close()

	// Bootstrap if needed
	if err := db.BootstrapDB(database); err != nil {
		return err
	}

	// Prompts
	artist := promptString("Artist: ")
	title := promptString("Title: ")
	manualYear := promptOptionalInt("Year (optional): ")
	formatCategory := promptValidFormat("Format category (CD, Vinyl, Tape, Digital): ")
	formatDetailInput := promptString("Format detail (optional): ")
	var formatDetail *string
	if formatDetailInput != "" {
		formatDetail = &formatDetailInput
	}

	// Insert release
	id, err := db.InsertRelease(database, artist, title, manualYear, nil)
	if err != nil {
		return err
	}

	// Insert ownership
	_, err = db.InsertOwnership(database, id, formatCategory, formatDetail, nil, nil, nil, nil, nil, false)
	if err != nil {
		return err
	}

	fmt.Println("Added release to collection.")
	return nil
}

func searchArtists(query string) ([]Artist, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/artist?query=artist:%s&fmt=json", query)
	resp, err := makeRequest(url)
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

func searchReleaseGroups(artistID string) ([]ReleaseGroup, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group?artist=%s&fmt=json", artistID)
	resp, err := makeRequest(url)
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

func makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "musiccat/0.1.0 (robertjamespeacock@gmail.com)")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}
	time.Sleep(1 * time.Second) // Rate limit
	return resp, nil
}

func init() {
	addCmd.Flags().Bool("manual", false, "Manually enter release details")
	addCmd.Flags().Bool("manual", false, "Manually enter release details")
	addCmd.Flags().Int("page-size", 40, "Number of releases per page")
	rootCmd.AddCommand(addCmd)
}
