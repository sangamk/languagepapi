//go:build ignore

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
	"google.golang.org/genai"
)

func main() {
	_ = godotenv.Load()
	
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal("Failed to create Gemini client:", err)
	}
	
	db, err := sql.Open("sqlite", "languagepapi.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get songs that need translation
	rows, err := db.Query(`
		SELECT DISTINCT s.id, s.title 
		FROM songs s
		JOIN song_lines sl ON s.id = sl.song_id
		WHERE sl.english_text = '' OR sl.english_text IS NULL
		ORDER BY s.title
	`)
	if err != nil {
		log.Fatal(err)
	}
	
	var songs []struct{ id int64; title string }
	for rows.Next() {
		var s struct{ id int64; title string }
		rows.Scan(&s.id, &s.title)
		songs = append(songs, s)
	}
	rows.Close()
	
	fmt.Printf("Found %d songs needing translation\n\n", len(songs))
	
	for _, song := range songs {
		fmt.Printf("Translating: %s... ", song.title)
		
		// Get all lines for this song
		lineRows, _ := db.Query(`
			SELECT id, spanish_text FROM song_lines 
			WHERE song_id = ? ORDER BY line_number
		`, song.id)
		
		var lines []struct{ id int64; text string }
		for lineRows.Next() {
			var l struct{ id int64; text string }
			lineRows.Scan(&l.id, &l.text)
			lines = append(lines, l)
		}
		lineRows.Close()
		
		if len(lines) == 0 {
			fmt.Println("no lines")
			continue
		}
		
		// Build translation prompt
		var spanishLines []string
		for _, l := range lines {
			spanishLines = append(spanishLines, l.text)
		}
		
		prompt := fmt.Sprintf(`Translate these Spanish song lyrics to English.
Return ONLY a JSON array of strings with the translations in the same order.
Keep translations natural and conversational.

Spanish lines:
%s

Return format: ["translation1", "translation2", ...]`, strings.Join(spanishLines, "\n"))

		result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), nil)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		
		response := result.Text()
		
		// Parse JSON
		start := strings.Index(response, "[")
		end := strings.LastIndex(response, "]")
		if start == -1 || end == -1 {
			fmt.Println("bad response format")
			continue
		}
		
		var translations []string
		if err := json.Unmarshal([]byte(response[start:end+1]), &translations); err != nil {
			fmt.Printf("parse error: %v\n", err)
			continue
		}
		
		// Update database
		updated := 0
		for i, l := range lines {
			if i < len(translations) && translations[i] != "" {
				db.Exec(`UPDATE song_lines SET english_text = ? WHERE id = ?`, translations[i], l.id)
				updated++
			}
		}
		
		fmt.Printf("OK (%d/%d lines)\n", updated, len(lines))
		time.Sleep(1 * time.Second) // Small delay between songs
	}
	
	fmt.Println("\nDone!")
}
