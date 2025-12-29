package repository

import (
	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// LogReview records a review event
func LogReview(log *models.ReviewLog) error {
	result, err := db.DB.Exec(`
		INSERT INTO review_logs (user_id, card_id, rating, elapsed_days, scheduled_days, review_duration_ms)
		VALUES (?, ?, ?, ?, ?, ?)
	`, log.UserID, log.CardID, log.Rating, log.ElapsedDays, log.ScheduledDays, log.ReviewDurationMs)
	if err != nil {
		return err
	}
	log.ID, err = result.LastInsertId()
	return err
}

// GetReviewHistory retrieves review history for a card
func GetReviewHistory(userID, cardID int64) ([]models.ReviewLog, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, card_id, rating, elapsed_days, scheduled_days, reviewed_at, review_duration_ms
		FROM review_logs
		WHERE user_id = ? AND card_id = ?
		ORDER BY reviewed_at DESC
	`, userID, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ReviewLog
	for rows.Next() {
		var l models.ReviewLog
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.CardID, &l.Rating,
			&l.ElapsedDays, &l.ScheduledDays, &l.ReviewedAt, &l.ReviewDurationMs,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// GetTotalReviewCount returns total reviews for a user
func GetTotalReviewCount(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM review_logs WHERE user_id = ?
	`, userID).Scan(&count)
	return count, err
}

// GetTodayReviewCount returns reviews done today
func GetTodayReviewCount(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM review_logs
		WHERE user_id = ? AND DATE(reviewed_at) = DATE('now')
	`, userID).Scan(&count)
	return count, err
}
