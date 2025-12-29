package repository

import (
	"database/sql"
	"time"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetUser retrieves a user by ID
func GetUser(id int64) (*models.User, error) {
	user := &models.User{}
	err := db.DB.QueryRow(`
		SELECT id, username, created_at, total_xp, current_streak, longest_streak, last_active_date
		FROM users WHERE id = ?
	`, id).Scan(
		&user.ID, &user.Username, &user.CreatedAt,
		&user.TotalXP, &user.CurrentStreak, &user.LongestStreak, &user.LastActiveDate,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserXP adds XP to a user's total
func UpdateUserXP(userID int64, xpToAdd int) error {
	_, err := db.DB.Exec(`
		UPDATE users SET total_xp = total_xp + ? WHERE id = ?
	`, xpToAdd, userID)
	return err
}

// UpdateStreak updates the user's streak based on activity
func UpdateStreak(userID int64) error {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Get current streak info
	var lastActive sql.NullString
	var currentStreak, longestStreak int
	err := db.DB.QueryRow(`
		SELECT last_active_date, current_streak, longest_streak
		FROM users WHERE id = ?
	`, userID).Scan(&lastActive, &currentStreak, &longestStreak)
	if err != nil {
		return err
	}

	var newStreak int
	if !lastActive.Valid {
		// First activity ever
		newStreak = 1
	} else if lastActive.String == today {
		// Already active today, no change
		return nil
	} else if lastActive.String == yesterday {
		// Consecutive day, increment streak
		newStreak = currentStreak + 1
	} else {
		// Streak broken, restart
		newStreak = 1
	}

	// Update longest streak if needed
	newLongest := longestStreak
	if newStreak > longestStreak {
		newLongest = newStreak
	}

	_, err = db.DB.Exec(`
		UPDATE users
		SET current_streak = ?, longest_streak = ?, last_active_date = ?
		WHERE id = ?
	`, newStreak, newLongest, today, userID)
	return err
}

// GetStreakInfo returns the user's streak data
func GetStreakInfo(userID int64) (*models.StreakInfo, error) {
	info := &models.StreakInfo{}
	var lastActive sql.NullString
	err := db.DB.QueryRow(`
		SELECT current_streak, longest_streak, last_active_date
		FROM users WHERE id = ?
	`, userID).Scan(&info.CurrentStreak, &info.LongestStreak, &lastActive)
	if err != nil {
		return nil, err
	}

	if lastActive.Valid {
		t, _ := time.Parse("2006-01-02", lastActive.String)
		info.LastActive = sql.NullTime{Time: t, Valid: true}
		info.IsActiveToday = lastActive.String == time.Now().Format("2006-01-02")
	}
	return info, nil
}
