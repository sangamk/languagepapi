package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"languagepapi/internal/db"
	"languagepapi/internal/handlers"
)

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

	// Routes
	http.HandleFunc("GET /", handlers.HandleHome)

	// Practice routes
	http.HandleFunc("GET /practice", handlers.HandlePractice)
	http.HandleFunc("GET /practice/card", handlers.HandlePracticeCard)
	http.HandleFunc("POST /practice/review", handlers.HandleReview)
	http.HandleFunc("POST /practice/skip", handlers.HandleSkip)
	http.HandleFunc("GET /practice/stats", handlers.HandlePracticeStats)

	// Words management
	http.HandleFunc("GET /words", handlers.HandleWords)
	http.HandleFunc("GET /add", handlers.HandleAddCard)
	http.HandleFunc("POST /add", handlers.HandleCreateCard)
	http.HandleFunc("GET /words/{id}/edit", handlers.HandleEditCard)
	http.HandleFunc("PUT /words/{id}", handlers.HandleUpdateCard)
	http.HandleFunc("DELETE /words/{id}", handlers.HandleDeleteCard)

	// AI generation routes
	http.HandleFunc("POST /words/{id}/generate-bridges", handlers.HandleGenerateBridges)
	http.HandleFunc("POST /words/{id}/generate-example", handlers.HandleGenerateExample)

	// Calendar / Stats
	http.HandleFunc("GET /calendar", handlers.HandleCalendar)

	// Settings
	http.HandleFunc("GET /settings", handlers.HandleSettings)
	http.HandleFunc("POST /settings", handlers.HandleSaveSettings)

	// Static files
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// getEnv returns environment variable or default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
