package repository

import (
	"database/sql"
	"time"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetJourney retrieves the active journey for a user
func GetJourney(userID int64) (*models.CurriculumJourney, error) {
	j := &models.CurriculumJourney{}
	err := db.DB.QueryRow(`
		SELECT id, user_id, start_date, is_active, created_at
		FROM curriculum_journey
		WHERE user_id = ? AND is_active = 1
	`, userID).Scan(&j.ID, &j.UserID, &j.StartDate, &j.IsActive, &j.CreatedAt)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// CreateJourney creates a new journey starting today
func CreateJourney(userID int64) (*models.CurriculumJourney, error) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	result, err := db.DB.Exec(`
		INSERT INTO curriculum_journey (user_id, start_date, is_active)
		VALUES (?, ?, 1)
		ON CONFLICT(user_id) DO UPDATE SET
			start_date = excluded.start_date,
			is_active = 1
	`, userID, startDate)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &models.CurriculumJourney{
		ID:        id,
		UserID:    userID,
		StartDate: startDate,
		IsActive:  true,
		CreatedAt: now,
	}, nil
}

// GetOrCreateJourney gets existing journey or creates a new one
func GetOrCreateJourney(userID int64) (*models.CurriculumJourney, error) {
	journey, err := GetJourney(userID)
	if err == sql.ErrNoRows {
		return CreateJourney(userID)
	}
	return journey, err
}

// GetTodayLessonSession retrieves today's lesson session if exists
func GetTodayLessonSession(userID int64) (*models.LessonSession, error) {
	today := time.Now().Format("2006-01-02")
	s := &models.LessonSession{}
	err := db.DB.QueryRow(`
		SELECT id, user_id, session_date, day_number, phase_id,
		       cards_reviewed, cards_correct, new_cards_learned, xp_earned,
		       completed_at, created_at
		FROM lesson_sessions
		WHERE user_id = ? AND session_date = ?
	`, userID, today).Scan(
		&s.ID, &s.UserID, &s.SessionDate, &s.DayNumber, &s.PhaseID,
		&s.CardsReviewed, &s.CardsCorrect, &s.NewCardsLearned, &s.XPEarned,
		&s.CompletedAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateLessonSession creates a new lesson session for today
func CreateLessonSession(userID int64, dayNumber, phaseID int) (*models.LessonSession, error) {
	today := time.Now().Format("2006-01-02")
	result, err := db.DB.Exec(`
		INSERT INTO lesson_sessions (user_id, session_date, day_number, phase_id)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, session_date) DO NOTHING
	`, userID, today, dayNumber, phaseID)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &models.LessonSession{
		ID:          id,
		UserID:      userID,
		SessionDate: time.Now(),
		DayNumber:   dayNumber,
		PhaseID:     phaseID,
	}, nil
}

// UpdateLessonSession updates session stats during practice
func UpdateLessonSession(sessionID int64, reviewed, correct, newLearned, xp int) error {
	_, err := db.DB.Exec(`
		UPDATE lesson_sessions
		SET cards_reviewed = cards_reviewed + ?,
		    cards_correct = cards_correct + ?,
		    new_cards_learned = new_cards_learned + ?,
		    xp_earned = xp_earned + ?
		WHERE id = ?
	`, reviewed, correct, newLearned, xp, sessionID)
	return err
}

// CompleteLessonSession marks a lesson session as complete
func CompleteLessonSession(sessionID int64) error {
	_, err := db.DB.Exec(`
		UPDATE lesson_sessions
		SET completed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, sessionID)
	return err
}

// IsTodayLessonComplete checks if today's lesson has been completed
func IsTodayLessonComplete(userID int64) (bool, *models.LessonSession, error) {
	session, err := GetTodayLessonSession(userID)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	return session.CompletedAt.Valid, session, nil
}

// GetNewCardsFromIslands retrieves new cards from specific islands
func GetNewCardsFromIslands(userID int64, islandIDs []int64, limit int) ([]models.CardWithProgress, error) {
	if len(islandIDs) == 0 {
		return nil, nil
	}

	// Build the IN clause dynamically
	query := `
		SELECT c.id, c.island_id, c.term, c.translation,
		       COALESCE(c.example_sentence, ''), COALESCE(c.notes, ''), COALESCE(c.audio_url, ''),
		       c.frequency_rank, c.created_at
		FROM cards c
		LEFT JOIN card_progress p ON c.id = p.card_id AND p.user_id = ?
		WHERE (p.id IS NULL OR p.state = 'new')
		  AND c.island_id IN (`

	args := []interface{}{userID}
	for i, id := range islandIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += `)
		ORDER BY c.frequency_rank ASC, c.id ASC
		LIMIT ?`
	args = append(args, limit)

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
			&cwp.ExampleSentence, &cwp.Notes, &cwp.AudioURL, &cwp.FrequencyRank, &cwp.CreatedAt,
		); err != nil {
			return nil, err
		}
		cwp.Progress = nil // New card, no progress yet
		cards = append(cards, cwp)
	}
	return cards, rows.Err()
}
