package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"languagepapi/internal/db"
	"languagepapi/internal/service"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Get config from environment
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "languagepapi.db"
	}

	// Initialize database
	if err := db.Init(dbPath); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Fetching lyrics for all songs from lrclib.net...")
	fmt.Println("This will also translate lyrics to English using Gemini (if GEMINI_API_KEY is set)")
	fmt.Println()

	lyricsService := service.NewLyricsService()
	success, failed, err := lyricsService.FetchLyricsForAllSongs()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Printf("Done! Fetched lyrics for %d songs, %d failed\n", success, failed)
}
