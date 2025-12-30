package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"languagepapi/internal/bridge"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// LyricsService handles fetching and parsing synced lyrics
type LyricsService struct {
	httpClient    *http.Client
	geminiService *bridge.GeminiService
}

// LRCLibResponse represents the response from lrclib.net API
type LRCLibResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

// LyricLine represents a parsed lyric line with timing
type LyricLine struct {
	StartTimeMs int
	EndTimeMs   int
	Text        string
}

// NewLyricsService creates a new lyrics service
func NewLyricsService() *LyricsService {
	gemini, _ := bridge.NewGeminiService(context.Background()) // May be nil if no API key
	return &LyricsService{
		httpClient:    &http.Client{},
		geminiService: gemini,
	}
}

// FetchLyrics fetches synced lyrics from lrclib.net
func (s *LyricsService) FetchLyrics(artist, track, album string, durationSec int) (*LRCLibResponse, error) {
	baseURL := "https://lrclib.net/api/get"
	params := url.Values{}
	params.Set("artist_name", artist)
	params.Set("track_name", track)
	if album != "" {
		params.Set("album_name", album)
	}
	if durationSec > 0 {
		params.Set("duration", strconv.Itoa(durationSec))
	}

	reqURL := baseURL + "?" + params.Encode()
	resp, err := s.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lyrics not found for %s - %s", artist, track)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lrclib API error: %s - %s", resp.Status, string(body))
	}

	var result LRCLibResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ParseSyncedLyrics parses LRC format lyrics into structured lines
// Format: [MM:SS.MS] lyric text
func (s *LyricsService) ParseSyncedLyrics(syncedLyrics string) ([]LyricLine, error) {
	if syncedLyrics == "" {
		return nil, fmt.Errorf("no synced lyrics provided")
	}

	// Regex to match [MM:SS.MS] or [MM:SS:MS] format
	lineRegex := regexp.MustCompile(`\[(\d{2}):(\d{2})[.:](\d{2,3})\]\s*(.*)`)

	lines := strings.Split(syncedLyrics, "\n")
	var result []LyricLine

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) != 5 {
			continue
		}

		minutes, _ := strconv.Atoi(matches[1])
		seconds, _ := strconv.Atoi(matches[2])
		ms := matches[3]
		// Handle both .XX and .XXX formats
		milliseconds, _ := strconv.Atoi(ms)
		if len(ms) == 2 {
			milliseconds *= 10 // Convert centiseconds to milliseconds
		}

		text := strings.TrimSpace(matches[4])
		if text == "" {
			continue // Skip empty lines
		}

		startTimeMs := (minutes*60+seconds)*1000 + milliseconds

		result = append(result, LyricLine{
			StartTimeMs: startTimeMs,
			Text:        text,
		})
	}

	// Calculate end times (each line ends when the next one starts)
	for i := 0; i < len(result); i++ {
		if i < len(result)-1 {
			result[i].EndTimeMs = result[i+1].StartTimeMs
		} else {
			// Last line: add 5 seconds as default duration
			result[i].EndTimeMs = result[i].StartTimeMs + 5000
		}
	}

	return result, nil
}

// TranslateLyrics translates Spanish lyrics to English using Gemini
func (s *LyricsService) TranslateLyrics(lines []LyricLine) ([]string, error) {
	if s.geminiService == nil {
		// Return empty translations if no Gemini service
		result := make([]string, len(lines))
		for i := range result {
			result[i] = ""
		}
		return result, nil
	}

	// Build a batch translation prompt
	var spanishLines []string
	for _, line := range lines {
		spanishLines = append(spanishLines, line.Text)
	}

	prompt := fmt.Sprintf(`Translate these Spanish song lyrics to English.
Return ONLY a JSON array of strings with the translations in the same order.
Keep translations natural and conversational.

Spanish lines:
%s

Return format: ["translation1", "translation2", ...]`, strings.Join(spanishLines, "\n"))

	response, err := s.geminiService.GenerateContent(context.Background(), prompt)
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	// Parse JSON array from response
	// Find JSON array in response
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("invalid translation response format")
	}

	jsonStr := response[start : end+1]
	var translations []string
	if err := json.Unmarshal([]byte(jsonStr), &translations); err != nil {
		return nil, fmt.Errorf("failed to parse translations: %w", err)
	}

	// Ensure we have the right number of translations
	if len(translations) != len(lines) {
		// Pad or truncate
		result := make([]string, len(lines))
		for i := range result {
			if i < len(translations) {
				result[i] = translations[i]
			} else {
				result[i] = ""
			}
		}
		return result, nil
	}

	return translations, nil
}

// FetchAndStoreLyrics fetches lyrics from lrclib and stores them in the database
func (s *LyricsService) FetchAndStoreLyrics(songID int64) error {
	// Get song details
	song, err := repository.GetSong(songID)
	if err != nil {
		return fmt.Errorf("song not found: %w", err)
	}

	// Check if lyrics already exist
	existingLines, _ := repository.GetSongLines(songID)
	if len(existingLines) > 0 {
		return fmt.Errorf("song already has %d lyrics lines", len(existingLines))
	}

	// Fetch from lrclib
	lrcResp, err := s.FetchLyrics(song.Artist, song.Title, song.Album, song.DurationSeconds)
	if err != nil {
		return err
	}

	if lrcResp.SyncedLyrics == "" {
		return fmt.Errorf("no synced lyrics available")
	}

	// Parse lyrics
	lines, err := s.ParseSyncedLyrics(lrcResp.SyncedLyrics)
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return fmt.Errorf("no lyrics lines parsed")
	}

	// Translate
	translations, err := s.TranslateLyrics(lines)
	if err != nil {
		// Continue without translations if it fails
		translations = make([]string, len(lines))
	}

	// Store in database
	for i, line := range lines {
		englishText := ""
		if i < len(translations) {
			englishText = translations[i]
		}

		songLine := &models.SongLine{
			SongID:      songID,
			LineNumber:  i + 1,
			StartTimeMs: line.StartTimeMs,
			EndTimeMs:   line.EndTimeMs,
			SpanishText: line.Text,
			EnglishText: englishText,
		}

		if err := repository.CreateSongLine(songLine); err != nil {
			return fmt.Errorf("failed to store line %d: %w", i+1, err)
		}
	}

	return nil
}

// FetchLyricsForAllSongs fetches lyrics for all songs that don't have them
func (s *LyricsService) FetchLyricsForAllSongs() (int, int, error) {
	songs, err := repository.GetAllSongs(0)
	if err != nil {
		return 0, 0, err
	}

	success := 0
	failed := 0

	for _, song := range songs {
		// Check if already has lyrics
		lines, _ := repository.GetSongLines(song.ID)
		if len(lines) > 0 {
			continue
		}

		err := s.FetchAndStoreLyrics(song.ID)
		if err != nil {
			fmt.Printf("Failed to fetch lyrics for %s: %v\n", song.Title, err)
			failed++
		} else {
			fmt.Printf("Fetched lyrics for %s\n", song.Title)
			success++
		}
	}

	return success, failed, nil
}
