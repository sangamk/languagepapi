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

	// Initialize database
	if err := db.Init(dbPath); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a new ServeMux to avoid conflicts with default mux
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /", handlers.HandleHome)

	// Practice routes
	mux.HandleFunc("GET /practice", handlers.HandlePractice)
	mux.HandleFunc("GET /practice/card", handlers.HandlePracticeCard)
	mux.HandleFunc("POST /practice/review", handlers.HandleReview)
	mux.HandleFunc("POST /practice/skip", handlers.HandleSkip)
	mux.HandleFunc("GET /practice/stats", handlers.HandlePracticeStats)

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

	// Settings
	mux.HandleFunc("GET /settings", handlers.HandleSettings)
	mux.HandleFunc("POST /settings", handlers.HandleSaveSettings)

	// Static files (embedded in binary)
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

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
