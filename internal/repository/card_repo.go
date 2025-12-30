package repository

import (
	"database/sql"
	"time"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetCard retrieves a card by ID
func GetCard(id int64) (*models.Card, error) {
	card := &models.Card{}
	err := db.DB.QueryRow(`
		SELECT id, island_id, term, translation,
		       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
		       frequency_rank, created_at
		FROM cards WHERE id = ?
	`, id).Scan(
		&card.ID, &card.IslandID, &card.Term, &card.Translation,
		&card.ExampleSentence, &card.Notes, &card.AudioURL, &card.FrequencyRank, &card.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return card, nil
}

// GetCardWithBridges retrieves a card with its bridges
func GetCardWithBridges(id int64) (*models.Card, error) {
	card, err := GetCard(id)
	if err != nil {
		return nil, err
	}

	rows, err := db.DB.Query(`
		SELECT id, card_id, bridge_type, bridge_content, explanation
		FROM bridges WHERE card_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b models.Bridge
		if err := rows.Scan(&b.ID, &b.CardID, &b.BridgeType, &b.BridgeContent, &b.Explanation); err != nil {
			return nil, err
		}
		card.Bridges = append(card.Bridges, b)
	}
	return card, rows.Err()
}

// GetCardsByIsland retrieves all cards for an island
func GetCardsByIsland(islandID int64) ([]models.Card, error) {
	rows, err := db.DB.Query(`
		SELECT id, island_id, term, translation,
		       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
		       frequency_rank, created_at
		FROM cards WHERE island_id = ?
		ORDER BY frequency_rank ASC, id ASC
	`, islandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(
			&c.ID, &c.IslandID, &c.Term, &c.Translation,
			&c.ExampleSentence, &c.Notes, &c.AudioURL, &c.FrequencyRank, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// CreateCard inserts a new card
func CreateCard(card *models.Card) error {
	// Default source to "curriculum" if not set
	source := card.Source
	if source == "" {
		source = "curriculum"
	}
	result, err := db.DB.Exec(`
		INSERT INTO cards (island_id, term, translation, example_sentence, notes, audio_url, frequency_rank, source, source_song_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, card.IslandID, card.Term, card.Translation, card.ExampleSentence, card.Notes, card.AudioURL, card.FrequencyRank, source, card.SourceSongID)
	if err != nil {
		return err
	}
	card.ID, err = result.LastInsertId()
	return err
}

// CreateBridge inserts a new bridge for a card
func CreateBridge(bridge *models.Bridge) error {
	result, err := db.DB.Exec(`
		INSERT INTO bridges (card_id, bridge_type, bridge_content, explanation)
		VALUES (?, ?, ?, ?)
	`, bridge.CardID, bridge.BridgeType, bridge.BridgeContent, bridge.Explanation)
	if err != nil {
		return err
	}
	bridge.ID, err = result.LastInsertId()
	return err
}

// DeleteCard removes a card and its bridges (cascades)
func DeleteCard(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM cards WHERE id = ?`, id)
	return err
}

// UpdateCard updates an existing card
func UpdateCard(card *models.Card) error {
	_, err := db.DB.Exec(`
		UPDATE cards SET
			island_id = ?,
			term = ?,
			translation = ?,
			example_sentence = ?,
			notes = ?,
			audio_url = ?
		WHERE id = ?
	`, card.IslandID, card.Term, card.Translation, card.ExampleSentence, card.Notes, card.AudioURL, card.ID)
	return err
}

// DeleteBridgesForCard removes all bridges for a card
func DeleteBridgesForCard(cardID int64) error {
	_, err := db.DB.Exec(`DELETE FROM bridges WHERE card_id = ?`, cardID)
	return err
}

// SearchCards searches for cards by term or translation
func SearchCards(query string, islandID int64) ([]models.Card, error) {
	searchPattern := "%" + query + "%"
	var rows *sql.Rows
	var err error

	if islandID > 0 {
		rows, err = db.DB.Query(`
			SELECT id, island_id, term, translation,
			       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
			       frequency_rank, created_at
			FROM cards
			WHERE island_id = ? AND (term LIKE ? OR translation LIKE ? OR example_sentence LIKE ?)
			ORDER BY frequency_rank ASC, id ASC
			LIMIT 100
		`, islandID, searchPattern, searchPattern, searchPattern)
	} else {
		rows, err = db.DB.Query(`
			SELECT id, island_id, term, translation,
			       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
			       frequency_rank, created_at
			FROM cards
			WHERE term LIKE ? OR translation LIKE ? OR example_sentence LIKE ?
			ORDER BY frequency_rank ASC, id ASC
			LIMIT 100
		`, searchPattern, searchPattern, searchPattern)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(
			&c.ID, &c.IslandID, &c.Term, &c.Translation,
			&c.ExampleSentence, &c.Notes, &c.AudioURL, &c.FrequencyRank, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// GetBridgesForCard retrieves all bridges for a card
func GetBridgesForCard(cardID int64) ([]models.Bridge, error) {
	rows, err := db.DB.Query(`
		SELECT id, card_id, bridge_type, bridge_content, COALESCE(explanation, '')
		FROM bridges WHERE card_id = ?
	`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bridges []models.Bridge
	for rows.Next() {
		var b models.Bridge
		if err := rows.Scan(&b.ID, &b.CardID, &b.BridgeType, &b.BridgeContent, &b.Explanation); err != nil {
			return nil, err
		}
		bridges = append(bridges, b)
	}
	return bridges, rows.Err()
}

// CountCards returns total card count
func CountCards() (int, error) {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM cards`).Scan(&count)
	return count, err
}

// GetRandomTranslations gets random translations for MCQ distractors
func GetRandomTranslations(excludeCardID int64, count int) ([]string, error) {
	rows, err := db.DB.Query(`
		SELECT translation FROM cards
		WHERE id != ?
		ORDER BY RANDOM()
		LIMIT ?
	`, excludeCardID, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}
	return translations, rows.Err()
}

// GetAllCards retrieves all cards with pagination
func GetAllCards(limit, offset int) ([]models.Card, error) {
	rows, err := db.DB.Query(`
		SELECT id, island_id, term, translation,
		       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
		       frequency_rank, created_at
		FROM cards
		ORDER BY frequency_rank ASC, id ASC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(
			&c.ID, &c.IslandID, &c.Term, &c.Translation,
			&c.ExampleSentence, &c.Notes, &c.AudioURL, &c.FrequencyRank, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// GetDueSongVocabCards fetches song vocabulary cards due for review
func GetDueSongVocabCards(userID int64, limit int) ([]models.CardWithProgress, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	rows, err := db.DB.Query(`
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, COALESCE(c.source, 'curriculum'), c.source_song_id, c.created_at,
		       p.id, p.user_id, p.card_id, p.stability, p.difficulty, p.elapsed_days, p.scheduled_days,
		       p.reps, p.lapses, p.state, p.due, p.last_review
		FROM cards c
		JOIN card_progress p ON c.id = p.card_id AND p.user_id = ?
		WHERE c.source = 'song'
		  AND p.state IN ('learning', 'review', 'relearning')
		  AND substr(p.due, 1, 19) <= ?
		ORDER BY p.due ASC
		LIMIT ?
	`, userID, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.CardWithProgress
	for rows.Next() {
		var cwp models.CardWithProgress
		var p models.CardProgress
		if err := rows.Scan(
			&cwp.ID, &cwp.IslandID, &cwp.Term, &cwp.Translation,
			&cwp.ExampleSentence, &cwp.Notes, &cwp.AudioURL,
			&cwp.FrequencyRank, &cwp.Source, &cwp.SourceSongID, &cwp.CreatedAt,
			&p.ID, &p.UserID, &p.CardID, &p.Stability, &p.Difficulty,
			&p.ElapsedDays, &p.ScheduledDays, &p.Reps, &p.Lapses,
			&p.State, &p.Due, &p.LastReview,
		); err != nil {
			return nil, err
		}
		cwp.Progress = &p
		cards = append(cards, cwp)
	}
	return cards, rows.Err()
}

// GetNewSongVocabCards fetches new (unlearned) song vocabulary cards
func GetNewSongVocabCards(userID int64, songIDs []int64, limit int) ([]models.CardWithProgress, error) {
	if len(songIDs) == 0 {
		return nil, nil
	}

	// Build placeholders for song IDs
	placeholders := "?"
	args := make([]interface{}, 0, len(songIDs)+2)
	args = append(args, songIDs[0])
	for i := 1; i < len(songIDs); i++ {
		placeholders += ",?"
		args = append(args, songIDs[i])
	}
	args = append(args, userID)
	args = append(args, limit)

	query := `
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, COALESCE(c.source, 'curriculum'), c.source_song_id, c.created_at
		FROM cards c
		WHERE c.source = 'song'
		  AND c.source_song_id IN (` + placeholders + `)
		  AND NOT EXISTS (
		      SELECT 1 FROM card_progress cp
		      WHERE cp.card_id = c.id AND cp.user_id = ?
		  )
		ORDER BY RANDOM()
		LIMIT ?
	`

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.CardWithProgress
	for rows.Next() {
		var cwp models.CardWithProgress
		if err := rows.Scan(
			&cwp.ID, &cwp.IslandID, &cwp.Term, &cwp.Translation,
			&cwp.ExampleSentence, &cwp.Notes, &cwp.AudioURL,
			&cwp.FrequencyRank, &cwp.Source, &cwp.SourceSongID, &cwp.CreatedAt,
		); err != nil {
			return nil, err
		}
		// No progress record yet - mark as new
		cwp.Progress = &models.CardProgress{State: models.StateNew}
		cards = append(cards, cwp)
	}
	return cards, rows.Err()
}

// GetSongTitleForCard retrieves the song title for a card from a song
func GetSongTitleForCard(cardID int64) (string, error) {
	var title string
	err := db.DB.QueryRow(`
		SELECT s.title FROM songs s
		JOIN cards c ON c.source_song_id = s.id
		WHERE c.id = ?
	`, cardID).Scan(&title)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return title, err
}
