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
	http.HandleFunc("GET /practice", handlers.HandlePractice)
	http.HandleFunc("GET /practice/card", handlers.HandlePracticeCard)
	http.HandleFunc("POST /practice/review", handlers.HandleReview)
	http.HandleFunc("GET /practice/stats", handlers.HandlePracticeStats)

	// Words management
	http.HandleFunc("GET /words", handlers.HandleWords)
	http.HandleFunc("GET /add", handlers.HandleAddCard)
	http.HandleFunc("POST /add", handlers.HandleCreateCard)
	http.HandleFunc("DELETE /words/{id}", handlers.HandleDeleteCard)

	// Calendar / Stats
	http.HandleFunc("GET /calendar", handlers.HandleCalendar)

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
