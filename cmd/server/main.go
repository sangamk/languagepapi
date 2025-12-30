package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"languagepapi/internal/db"
	"languagepapi/internal/handlers"
)

//go:embed static
var staticFiles embed.FS

func main() {
	// Load .env file (ignore error if not found - use system env vars)
	_ = godotenv.Load()

	// Get config from environment
	dbPath := getEnv("DB_PATH", "languagepapi.db")
	port := getEnv("PORT", "8080")
	songsPath := getEnv("SONGS_PATH", "./songs")

	// Initialize database
	if err := db.Init(dbPath); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a new ServeMux to avoid conflicts with default mux
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /", handlers.HandleHome)

	// Practice routes (legacy - kept for "extra practice")
	mux.HandleFunc("GET /practice", handlers.HandlePractice)
	mux.HandleFunc("GET /practice/card", handlers.HandlePracticeCard)
	mux.HandleFunc("POST /practice/review", handlers.HandleReview)
	mux.HandleFunc("POST /practice/skip", handlers.HandleSkip)
	mux.HandleFunc("GET /practice/stats", handlers.HandlePracticeStats)

	// Daily lesson routes (new journey mode)
	mux.HandleFunc("GET /lesson/start", handlers.HandleLessonStart)
	mux.HandleFunc("POST /lesson/review", handlers.HandleLessonReview)
	mux.HandleFunc("POST /lesson/skip", handlers.HandleLessonSkip)

	// Words management
	mux.HandleFunc("GET /words", handlers.HandleWords)
	mux.HandleFunc("GET /add", handlers.HandleAddCard)
	mux.HandleFunc("POST /add", handlers.HandleCreateCard)
	mux.HandleFunc("GET /words/{id}/edit", handlers.HandleEditCard)
	mux.HandleFunc("PUT /words/{id}", handlers.HandleUpdateCard)
	mux.HandleFunc("DELETE /words/{id}", handlers.HandleDeleteCard)

	// AI generation routes
	mux.HandleFunc("POST /words/{id}/generate-bridges", handlers.HandleGenerateBridges)
	mux.HandleFunc("POST /words/{id}/generate-example", handlers.HandleGenerateExample)

	// Calendar / Stats
	mux.HandleFunc("GET /calendar", handlers.HandleCalendar)

	// Progress Overview
	mux.HandleFunc("GET /progress", handlers.HandleProgress)

	// Settings
	mux.HandleFunc("GET /settings", handlers.HandleSettings)
	mux.HandleFunc("POST /settings", handlers.HandleSaveSettings)

	// Grammar
	mux.HandleFunc("GET /grammar", handlers.HandleGrammar)
	mux.HandleFunc("GET /grammar/{rule_key}", handlers.HandleGrammarDetail)
	mux.HandleFunc("GET /grammar/card/{card_id}", handlers.HandleGrammarForCard)

	// Song lessons
	mux.HandleFunc("GET /songs", handlers.HandleSongHome)
	mux.HandleFunc("GET /songs/{id}", handlers.HandleSongDetail)
	mux.HandleFunc("GET /songs/{id}/start", handlers.HandleSongStart)
	mux.HandleFunc("POST /songs/{id}/vocab-review", handlers.HandleSongVocabReview)
	mux.HandleFunc("POST /songs/{id}/next-phase", handlers.HandleSongNextPhase)
	mux.HandleFunc("POST /songs/{id}/next-line", handlers.HandleSongNextLine)
	mux.HandleFunc("POST /songs/{id}/skip-line", handlers.HandleSongSkipLine)
	mux.HandleFunc("POST /songs/{id}/submit-blank", handlers.HandleSongBlankSubmit)
	mux.HandleFunc("POST /songs/{id}/complete", handlers.HandleSongComplete)
	mux.HandleFunc("POST /songs/{id}/fetch-lyrics", handlers.HandleFetchLyrics)

	// Static files (embedded in binary)
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Songs served from filesystem (not embedded - keeps binary small)
	mux.Handle("GET /audio/", http.StripPrefix("/audio/", http.FileServer(http.Dir(songsPath))))

	// Album art extracted from audio file metadata
	handlers.SongsPath = songsPath
	mux.HandleFunc("GET /audio/cover/{filename}", handlers.HandleAlbumArt)

	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// getEnv returns environment variable or default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
