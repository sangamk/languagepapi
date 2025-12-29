package repository

import (
	"time"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetOrCreateToday retrieves or creates today's daily log
func GetOrCreateToday(userID int64) (*models.DailyLog, error) {
	today := time.Now().Format("2006-01-02")

	// Try to get existing log
	log := &models.DailyLog{}
	err := db.DB.QueryRow(`
		SELECT id, user_id, date, xp_earned, cards_reviewed, cards_correct, minutes_active, new_cards_added
		FROM daily_logs WHERE user_id = ? AND date = ?
	`, userID, today).Scan(
		&log.ID, &log.UserID, &log.Date, &log.XPEarned,
		&log.CardsReviewed, &log.CardsCorrect, &log.MinutesActive, &log.NewCardsAdded,
	)
	if err == nil {
		return log, nil
	}

	// Create new log for today
	result, err := db.DB.Exec(`
		INSERT INTO daily_logs (user_id, date) VALUES (?, ?)
	`, userID, today)
	if err != nil {
		return nil, err
	}

	log.ID, _ = result.LastInsertId()
	log.UserID = userID
	log.Date, _ = time.Parse("2006-01-02", today)
	return log, nil
}

// UpdateDailyLog updates an existing daily log
func UpdateDailyLog(log *models.DailyLog) error {
	_, err := db.DB.Exec(`
		UPDATE daily_logs
		SET xp_earned = ?, cards_reviewed = ?, cards_correct = ?, minutes_active = ?, new_cards_added = ?
		WHERE id = ?
	`, log.XPEarned, log.CardsReviewed, log.CardsCorrect, log.MinutesActive, log.NewCardsAdded, log.ID)
	return err
}

// IncrementDailyStats increments daily counters
func IncrementDailyStats(userID int64, xp, reviewed, correct, newCards int) error {
	today := time.Now().Format("2006-01-02")
	_, err := db.DB.Exec(`
		INSERT INTO daily_logs (user_id, date, xp_earned, cards_reviewed, cards_correct, new_cards_added)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, date) DO UPDATE SET
			xp_earned = xp_earned + excluded.xp_earned,
			cards_reviewed = cards_reviewed + excluded.cards_reviewed,
			cards_correct = cards_correct + excluded.cards_correct,
			new_cards_added = new_cards_added + excluded.new_cards_added
	`, userID, today, xp, reviewed, correct, newCards)
	return err
}

// GetHeatMapData retrieves daily logs for the heat map
func GetHeatMapData(userID int64, days int) ([]models.DailyLog, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	rows, err := db.DB.Query(`
		SELECT id, user_id, date, xp_earned, cards_reviewed, cards_correct, minutes_active, new_cards_added
		FROM daily_logs
		WHERE user_id = ? AND date >= ?
		ORDER BY date ASC
	`, userID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.DailyLog
	for rows.Next() {
		var l models.DailyLog
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.Date, &l.XPEarned,
			&l.CardsReviewed, &l.CardsCorrect, &l.MinutesActive, &l.NewCardsAdded,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// GetTodayStats retrieves today's statistics
func GetTodayStats(userID int64) (*models.TodayStats, error) {
	stats := &models.TodayStats{}
	today := time.Now().Format("2006-01-02")

	// Get daily log stats
	db.DB.QueryRow(`
		SELECT COALESCE(xp_earned, 0), COALESCE(cards_reviewed, 0), COALESCE(cards_correct, 0)
		FROM daily_logs WHERE user_id = ? AND date = ?
	`, userID, today).Scan(&stats.XPEarned, &stats.CardsReviewed, &stats.CardsCorrect)

	// Get due count
	stats.DueCount, _ = CountDueCards(userID)

	// Get new count
	stats.NewCount, _ = CountNewCards(userID)

	return stats, nil
}
