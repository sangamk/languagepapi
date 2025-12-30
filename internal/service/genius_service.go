package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// GeniusService handles interactions with the Genius API
type GeniusService struct {
	accessToken string
	httpClient  *http.Client
}

// GeniusSearchResult represents a search hit from Genius
type GeniusSearchResult struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Artist       string `json:"primary_artist_name"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"song_art_image_thumbnail_url"`
}

// GeniusSongDetail represents detailed song info from Genius
type GeniusSongDetail struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Artist       string `json:"primary_artist_name"`
	Album        string `json:"album_name"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"song_art_image_url"`
}

// NewGeniusService creates a new Genius API service
func NewGeniusService() (*GeniusService, error) {
	token := os.Getenv("GENIUS_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GENIUS_ACCESS_TOKEN environment variable not set")
	}

	return &GeniusService{
		accessToken: token,
		httpClient:  &http.Client{},
	}, nil
}

// Search searches for songs on Genius
func (s *GeniusService) Search(query string) ([]GeniusSearchResult, error) {
	endpoint := fmt.Sprintf("https://api.genius.com/search?q=%s", url.QueryEscape(query))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("genius API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Response struct {
			Hits []struct {
				Result struct {
					ID                         int    `json:"id"`
					Title                      string `json:"title"`
					PrimaryArtist              struct {
						Name string `json:"name"`
					} `json:"primary_artist"`
					URL                        string `json:"url"`
					SongArtImageThumbnailURL   string `json:"song_art_image_thumbnail_url"`
				} `json:"result"`
			} `json:"hits"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []GeniusSearchResult
	for _, hit := range result.Response.Hits {
		results = append(results, GeniusSearchResult{
			ID:           hit.Result.ID,
			Title:        hit.Result.Title,
			Artist:       hit.Result.PrimaryArtist.Name,
			URL:          hit.Result.URL,
			ThumbnailURL: hit.Result.SongArtImageThumbnailURL,
		})
	}

	return results, nil
}

// GetSong gets detailed song information from Genius
func (s *GeniusService) GetSong(geniusID int) (*GeniusSongDetail, error) {
	endpoint := fmt.Sprintf("https://api.genius.com/songs/%d", geniusID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("genius API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Response struct {
			Song struct {
				ID            int    `json:"id"`
				Title         string `json:"title"`
				PrimaryArtist struct {
					Name string `json:"name"`
				} `json:"primary_artist"`
				Album struct {
					Name string `json:"name"`
				} `json:"album"`
				URL              string `json:"url"`
				SongArtImageURL  string `json:"song_art_image_url"`
			} `json:"song"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	song := result.Response.Song
	albumName := ""
	if song.Album.Name != "" {
		albumName = song.Album.Name
	}

	return &GeniusSongDetail{
		ID:           song.ID,
		Title:        song.Title,
		Artist:       song.PrimaryArtist.Name,
		Album:        albumName,
		URL:          song.URL,
		ThumbnailURL: song.SongArtImageURL,
	}, nil
}

// SearchAndLinkSong searches Genius for a song and updates the database record
func (s *GeniusService) SearchAndLinkSong(songID int64, searchQuery string) error {
	results, err := s.Search(searchQuery)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return fmt.Errorf("no results found for: %s", searchQuery)
	}

	// Get the first result's details
	geniusSong, err := s.GetSong(results[0].ID)
	if err != nil {
		return err
	}

	// Update the song in our database
	return repository.UpdateSongGeniusInfo(songID, geniusSong.ID, geniusSong.Album, geniusSong.ThumbnailURL)
}

// ParseFilenameToSong parses a filename like "01. Moscow Mule.mp3" into title
func ParseFilenameToSong(filename string) (trackNum int, title string) {
	// Remove extension
	name := strings.TrimSuffix(filename, ".mp3")
	name = strings.TrimSuffix(name, ".flac")
	name = strings.TrimSuffix(name, ".m4a")

	// Try to parse track number: "01. Title" or "1. Title"
	re := regexp.MustCompile(`^(\d+)\.\s*(.+)$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) == 3 {
		fmt.Sscanf(matches[1], "%d", &trackNum)
		title = strings.TrimSpace(matches[2])
	} else {
		title = name
	}

	return trackNum, title
}

// CreateSongFromFile creates a song record from an audio file
func CreateSongFromFile(filename, artist, album string, difficulty int) (*models.Song, error) {
	_, title := ParseFilenameToSong(filename)

	song := &models.Song{
		Title:      title,
		Artist:     artist,
		Album:      album,
		Difficulty: difficulty,
		AudioPath:  filename,
	}

	if err := repository.CreateSong(song); err != nil {
		return nil, err
	}

	return song, nil
}

// GetSongByAudioPath finds a song by its audio path
func GetSongByAudioPath(audioPath string) (*models.Song, error) {
	return repository.GetSongByAudioPath(audioPath)
}
