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
	// Use substr to extract YYYY-MM-DD HH:MM:SS for consistent comparison
	// This works regardless of the timezone/monotonic clock suffix in stored dates
	now := time.Now().Format("2006-01-02 15:04:05")
	rows, err := db.DB.Query(`
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, c.created_at,
		       p.id, p.user_id, p.card_id, p.stability, p.difficulty, p.elapsed_days, p.scheduled_days,
		       p.reps, p.lapses, p.state, p.due, p.last_review
		FROM cards c
		INNER JOIN card_progress p ON c.id = p.card_id
		WHERE p.user_id = ? AND substr(p.due, 1, 19) <= ? AND p.state IN ('learning', 'review', 'relearning')
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
		ORDER BY RANDOM()
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
	now := time.Now().Format("2006-01-02 15:04:05")
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*)
		FROM card_progress
		WHERE user_id = ? AND substr(due, 1, 19) <= ? AND state IN ('learning', 'review', 'relearning')
	`, userID, now).Scan(&count)
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
	now := time.Now().Format("2006-01-02 15:04:05")
	if err = db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress
		WHERE user_id = ? AND substr(due, 1, 19) <= ? AND state IN ('learning', 'review', 'relearning')
	`, userID, now).Scan(&due); err != nil {
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

// GetLearnedCardsByIsland returns cards grouped by island that have been studied
func GetLearnedCardsByIsland(userID int64) ([]models.IslandProgress, error) {
	rows, err := db.DB.Query(`
		SELECT
			i.id, i.name, i.icon,
			COUNT(DISTINCT c.id) as total_cards,
			COUNT(DISTINCT CASE WHEN p.state IN ('learning', 'review', 'relearning') THEN c.id END) as learned_cards,
			COUNT(DISTINCT CASE WHEN p.stability > 30 AND p.reps >= 5 THEN c.id END) as mastered_cards
		FROM islands i
		LEFT JOIN cards c ON c.island_id = i.id
		LEFT JOIN card_progress p ON c.id = p.card_id AND p.user_id = ?
		GROUP BY i.id
		ORDER BY i.sort_order
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var islands []models.IslandProgress
	for rows.Next() {
		var ip models.IslandProgress
		if err := rows.Scan(&ip.IslandID, &ip.IslandName, &ip.IslandIcon, &ip.TotalCards, &ip.LearnedCards, &ip.MasteredCards); err != nil {
			return nil, err
		}
		islands = append(islands, ip)
	}
	return islands, rows.Err()
}

// GetWordsByState returns cards with their progress state
func GetWordsByState(userID int64, state models.CardState, limit int) ([]models.CardWithProgress, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, c.created_at,
		       p.id, p.user_id, p.card_id, p.stability, p.difficulty, p.elapsed_days, p.scheduled_days,
		       p.reps, p.lapses, p.state, p.due, p.last_review
		FROM cards c
		INNER JOIN card_progress p ON c.id = p.card_id
		WHERE p.user_id = ? AND p.state = ?
		ORDER BY p.last_review DESC
		LIMIT ?
	`, userID, state, limit)
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

// GetRecentlyLearnedWords returns the most recently learned words
func GetRecentlyLearnedWords(userID int64, limit int) ([]models.CardWithProgress, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, c.created_at,
		       p.id, p.user_id, p.card_id, p.stability, p.difficulty, p.elapsed_days, p.scheduled_days,
		       p.reps, p.lapses, p.state, p.due, p.last_review
		FROM cards c
		INNER JOIN card_progress p ON c.id = p.card_id
		WHERE p.user_id = ? AND p.state != 'new'
		ORDER BY p.last_review DESC
		LIMIT ?
	`, userID, limit)
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

// GetProgressOverviewStats returns aggregate stats for the overview page
func GetProgressOverviewStats(userID int64) (*models.ProgressOverviewStats, error) {
	stats := &models.ProgressOverviewStats{}

	// Total cards in system
	db.DB.QueryRow(`SELECT COUNT(*) FROM cards`).Scan(&stats.TotalCards)

	// Cards by state
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress WHERE user_id = ? AND state = 'learning'
	`, userID).Scan(&stats.LearningCount)

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress WHERE user_id = ? AND state = 'review'
	`, userID).Scan(&stats.ReviewCount)

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress WHERE user_id = ? AND state = 'relearning'
	`, userID).Scan(&stats.RelearningCount)

	// Mastered (high stability)
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress WHERE user_id = ? AND stability > 30 AND reps >= 5
	`, userID).Scan(&stats.MasteredCount)

	// Total learned (not new)
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM card_progress WHERE user_id = ? AND state != 'new'
	`, userID).Scan(&stats.LearnedCount)

	// Total reviews
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(reps), 0) FROM card_progress WHERE user_id = ?
	`, userID).Scan(&stats.TotalReviews)

	return stats, nil
}

// GetRecentLessonSessions returns the most recent lesson sessions
func GetRecentLessonSessions(userID int64, limit int) ([]models.LessonSession, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, session_date, day_number, phase_id,
		       cards_reviewed, cards_correct, new_cards_learned, xp_earned,
		       completed_at, created_at
		FROM lesson_sessions
		WHERE user_id = ? AND completed_at IS NOT NULL
		ORDER BY session_date DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.LessonSession
	for rows.Next() {
		var s models.LessonSession
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.SessionDate, &s.DayNumber, &s.PhaseID,
			&s.CardsReviewed, &s.CardsCorrect, &s.NewCardsLearned, &s.XPEarned,
			&s.CompletedAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
