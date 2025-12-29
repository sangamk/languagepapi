package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// SpanishWord represents a parsed word entry
type SpanishWord struct {
	Rank        int
	Term        string
	Translation string
	Notes       string
}

func main() {
	// Open database
	db, err := sql.Open("sqlite", "languagepapi.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Enable foreign keys
	db.Exec("PRAGMA foreign_keys = ON")

	// Parse spanish.json
	words, err := parseSpanishJSON("spanish.json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed %d words\n", len(words))

	// Insert words into database
	inserted := 0
	skipped := 0

	for _, w := range words {
		// Determine island based on frequency rank
		islandID := getIslandForRank(w.Rank)

		// Check if word already exists
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM cards WHERE term = ?", w.Term).Scan(&exists)
		if exists > 0 {
			skipped++
			continue
		}

		// Insert card
		_, err := db.Exec(`
			INSERT INTO cards (island_id, term, translation, notes, frequency_rank)
			VALUES (?, ?, ?, ?, ?)
		`, islandID, w.Term, w.Translation, w.Notes, w.Rank)

		if err != nil {
			log.Printf("Error inserting %s: %v", w.Term, err)
			continue
		}
		inserted++
	}

	fmt.Printf("Inserted: %d, Skipped (duplicates): %d\n", inserted, skipped)
}

// getIslandForRank returns the appropriate island ID based on word frequency
func getIslandForRank(rank int) int {
	switch {
	case rank <= 100:
		return 1 // Core Essentials
	case rank <= 250:
		return 2 // Common Words
	case rank <= 500:
		return 3 // Expanding Vocabulary
	default:
		return 4 // Advanced Vocabulary
	}
}

// parseSpanishJSON parses the malformed spanish.json file
func parseSpanishJSON(filename string) ([]SpanishWord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []SpanishWord
	scanner := bufio.NewScanner(file)

	// Read entire file
	var content strings.Builder
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	text := content.String()

	// Pattern to match entries like {"1", "term", "translation"} or {"1", "term - translation"}
	// The JSON is malformed with string values as object "keys"

	// First, let's try to extract entries between { and }
	entryRegex := regexp.MustCompile(`\{([^}]+)\}`)
	matches := entryRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		entry := match[1]
		word := parseEntry(entry)
		if word != nil {
			words = append(words, *word)
		}
	}

	return words, nil
}

// parseEntry parses a single entry from the malformed JSON
func parseEntry(entry string) *SpanishWord {
	// Clean up the entry
	entry = strings.TrimSpace(entry)

	// Split by comma, but be careful of commas inside quotes
	parts := splitQuotedString(entry)

	if len(parts) < 2 {
		return nil
	}

	// First part should be the rank
	rankStr := cleanString(parts[0])
	rank, err := strconv.Atoi(rankStr)
	if err != nil {
		return nil
	}

	var term, translation, notes string

	if len(parts) >= 3 {
		// Format: "rank", "term", "translation"
		term = cleanString(parts[1])
		translation = cleanString(parts[2])
		if len(parts) > 3 {
			notes = cleanString(strings.Join(parts[3:], ", "))
		}
	} else if len(parts) == 2 {
		// Format: "rank", "term - translation" or "rank", "term  -  translation"
		second := cleanString(parts[1])

		// Try to split on " - " or "  -  "
		if idx := strings.Index(second, "  -  "); idx > 0 {
			term = strings.TrimSpace(second[:idx])
			translation = strings.TrimSpace(second[idx+5:])
		} else if idx := strings.Index(second, " - "); idx > 0 {
			term = strings.TrimSpace(second[:idx])
			translation = strings.TrimSpace(second[idx+3:])
		} else {
			// No separator, treat whole thing as term
			term = second
			translation = ""
		}
	}

	// Clean up term (remove any trailing notes in brackets or dashes)
	if idx := strings.Index(term, "  -  "); idx > 0 {
		remaining := term[idx+5:]
		term = strings.TrimSpace(term[:idx])
		if translation == "" {
			translation = remaining
		}
	}

	if term == "" {
		return nil
	}

	// Extract notes from translation if present (stuff in brackets)
	if strings.Contains(translation, "[") {
		// Keep it in translation but also add to notes
		notes = translation
	}

	return &SpanishWord{
		Rank:        rank,
		Term:        term,
		Translation: translation,
		Notes:       notes,
	}
}

// splitQuotedString splits a string by comma, respecting quoted sections
func splitQuotedString(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
		} else if c == ',' && !inQuotes {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// cleanString removes quotes and trims whitespace
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	s = strings.TrimSpace(s)
	return s
}
