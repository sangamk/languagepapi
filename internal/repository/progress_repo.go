package repository

import (
	"database/sql"
	"time"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetProgress retrieves progress for a specific user+card
func GetProgress(userID, cardID int64) (*models.CardProgress, error) {
	p := &models.CardProgress{}
	err := db.DB.QueryRow(`
		SELECT id, user_id, card_id, stability, difficulty, elapsed_days, scheduled_days,
		       reps, lapses, state, due, last_review
		FROM card_progress WHERE user_id = ? AND card_id = ?
	`, userID, cardID).Scan(
		&p.ID, &p.UserID, &p.CardID, &p.Stability, &p.Difficulty,
		&p.ElapsedDays, &p.ScheduledDays, &p.Reps, &p.Lapses,
		&p.State, &p.Due, &p.LastReview,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// UpsertProgress creates or updates card progress
func UpsertProgress(p *models.CardProgress) error {
	result, err := db.DB.Exec(`
		INSERT INTO card_progress (user_id, card_id, stability, difficulty, elapsed_days, scheduled_days,
		                           reps, lapses, state, due, last_review)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, card_id) DO UPDATE SET
			stability = excluded.stability,
			difficulty = excluded.difficulty,
			elapsed_days = excluded.elapsed_days,
			scheduled_days = excluded.scheduled_days,
			reps = excluded.reps,
			lapses = excluded.lapses,
			state = excluded.state,
			due = excluded.due,
			last_review = excluded.last_review
	`, p.UserID, p.CardID, p.Stability, p.Difficulty, p.ElapsedDays, p.ScheduledDays,
		p.Reps, p.Lapses, p.State, p.Due, p.LastReview)
	if err != nil {
		return err
	}
	if p.ID == 0 {
		p.ID, _ = result.LastInsertId()
	}
	return nil
}

// GetDueCards returns cards due for review
func GetDueCards(userID int64, limit int) ([]models.CardWithProgress, error) {
	now := time.Now()
	rows, err := db.DB.Query(`
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, c.created_at,
		       p.id, p.user_id, p.card_id, p.stability, p.difficulty, p.elapsed_days, p.scheduled_days,
		       p.reps, p.lapses, p.state, p.due, p.last_review
		FROM cards c
		INNER JOIN card_progress p ON c.id = p.card_id
		WHERE p.user_id = ? AND p.due <= ? AND p.state IN ('learning', 'review', 'relearning')
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
			&cwp.ExampleSentence, &cwp.Notes, &cwp.AudioURL, &cwp.FrequencyRank, &cwp.CreatedAt,
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

// GetNewCards returns cards that haven't been studied yet
func GetNewCards(userID int64, limit int) ([]models.CardWithProgress, error) {
	// Get cards with no progress record OR with state='new'
	rows, err := db.DB.Query(`
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, c.created_at
		FROM cards c
		LEFT JOIN card_progress p ON c.id = p.card_id AND p.user_id = ?
		WHERE p.id IS NULL OR p.state = 'new'
		ORDER BY c.frequency_rank ASC, c.id ASC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.CardWithProgress
	for rows.Next() {
		var cwp models.CardWithProgress
		if err := rows.Scan(
			&cwp.ID, &cwp.IslandID, &cwp.Term, &cwp.Translation,
			&cwp.ExampleSentence, &cwp.Notes, &cwp.AudioURL, &cwp.FrequencyRank, &cwp.CreatedAt,
		); err != nil {
			return nil, err
		}
		// New card, no progress yet
		cwp.Progress = nil
		cards = append(cards, cwp)
	}
	return cards, rows.Err()
}

// CountDueCards returns the number of cards due for review
func CountDueCards(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*)
		FROM card_progress
		WHERE user_id = ? AND due <= ? AND state IN ('learning', 'review', 'relearning')
	`, userID, time.Now()).Scan(&count)
	return count, err
}

// CountNewCards returns the number of new cards not yet studied
func CountNewCards(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*)
		FROM cards c
		LEFT JOIN card_progress p ON c.id = p.card_id AND p.user_id = ?
		WHERE p.id IS NULL OR p.state = 'new'
	`, userID).Scan(&count)
	return count, err
}

// GetCardProgressStats returns stats for a user
func GetCardProgressStats(userID int64) (total, learned, due, mastered int, err error) {
	// Total cards
	if err = db.DB.QueryRow(`SELECT COUNT(*) FROM cards`).Scan(&total); err != nil {
		return
	}

	// Learned (has progress, not new)
	if err = db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress WHERE user_id = ? AND state != 'new'
	`, userID).Scan(&learned); err != nil {
		return
	}

	// Due
	if err = db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress
		WHERE user_id = ? AND due <= ? AND state IN ('learning', 'review', 'relearning')
	`, userID, time.Now()).Scan(&due); err != nil {
		return
	}

	// Mastered (stability > 30 days, high reps)
	if err = db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress
		WHERE user_id = ? AND stability > 30 AND reps >= 5
	`, userID).Scan(&mastered); err != nil {
		return
	}

	return
}

// InitializeProgressForCard creates a new progress record for a card
func InitializeProgressForCard(userID, cardID int64) (*models.CardProgress, error) {
	p := &models.CardProgress{
		UserID:     userID,
		CardID:     cardID,
		Stability:  0,
		Difficulty: 0,
		State:      models.StateNew,
		Due:        sql.NullTime{Time: time.Now(), Valid: true},
	}
	if err := UpsertProgress(p); err != nil {
		return nil, err
	}
	return p, nil
}
