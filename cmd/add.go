package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"musiccat/internal/config"
	"musiccat/internal/db"

	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		manual, _ := cmd.Flags().GetBool("manual")
		if manual {
			if len(args) != 0 {
				return fmt.Errorf("--manual takes no arguments")
			}
			return addManual()
		}
		if len(args) == 0 {
			return fmt.Errorf("requires artist name argument")
		}
		artistName := strings.Join(args, " ")
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
	totalAdded := 0
	for _, item := range selectedItems {
		rg := releaseGroups[item.index-1]
		var releaseYear *int = parseYear(rg.FirstReleaseDate)
		releaseID, err := db.InsertRelease(database, selectedArtist.Name, rg.Title, releaseYear, &rg.ID)
		if err != nil {
			return err
		}

		// Insert multiple ownership entries if quantity > 1
		for i := 0; i < item.quantity; i++ {
			var notes *string
			if item.quantity > 1 && item.notes != "" {
				notes = &item.notes
			}

			_, err = db.InsertOwnership(database, releaseID, cfg.CurrentFormat, nil, nil, nil, nil, notes, nil, item.promo)
			if err != nil {
				return err
			}
			totalAdded++
		}
	}

	fmt.Printf("Added %d ownership entries to collection.\n", totalAdded)
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
	formatCategory := promptValidFormat("Format category (CD, Vinyl, Cassette): ")
	formatDetailInput := promptFormatDetail(formatCategory)
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
	encodedQuery := url.QueryEscape(query)
	reqURL := fmt.Sprintf("https://musicbrainz.org/ws/2/artist?query=artist:%s&fmt=json", encodedQuery)
	resp, err := makeRequest(reqURL)
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
	// Request up to 100 release groups (MusicBrainz max)
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group?artist=%s&limit=100&fmt=json", artistID)
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

// createMusicBrainzClient creates an HTTP client configured for IPv4-only connections
func createMusicBrainzClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Force IPv4 by replacing "tcp" with "tcp4"
			return dialer.DialContext(ctx, "tcp4", addr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

func makeRequest(url string) (*http.Response, error) {
	const maxRetries = 3
	var lastErr error

	client := createMusicBrainzClient()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "musiccat/0.1 (robertjamespeacock@gmail.com)")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				// Exponential backoff: 1s, 2s
				waitTime := time.Duration(attempt) * time.Second
				fmt.Printf("Request failed (attempt %d/%d), retrying in %v...\n", attempt, maxRetries, waitTime)
				time.Sleep(waitTime)
				continue
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("API error: %s", resp.Status)
			if attempt < maxRetries {
				waitTime := time.Duration(attempt) * time.Second
				fmt.Printf("API returned %s (attempt %d/%d), retrying in %v...\n", resp.Status, attempt, maxRetries, waitTime)
				time.Sleep(waitTime)
				continue
			}
			return nil, lastErr
		}

		// Success - apply rate limiting before returning
		time.Sleep(1 * time.Second)
		return resp, nil
	}

	return nil, lastErr
}

func init() {
	addCmd.Flags().Bool("manual", false, "Manually enter release details")
	addCmd.Flags().Int("page-size", 50, "Number of releases per page")
	rootCmd.AddCommand(addCmd)
}
