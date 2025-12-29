package repository

import (
	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetAllAchievements retrieves all defined achievements
func GetAllAchievements() ([]models.Achievement, error) {
	rows, err := db.DB.Query(`
		SELECT id, name, description, icon, xp_reward, condition_type, condition_value
		FROM achievements ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []models.Achievement
	for rows.Next() {
		var a models.Achievement
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Icon, &a.XPReward, &a.ConditionType, &a.ConditionValue); err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}
	return achievements, rows.Err()
}

// GetUserAchievements retrieves achievements earned by a user
func GetUserAchievements(userID int64) ([]models.AchievementWithStatus, error) {
	rows, err := db.DB.Query(`
		SELECT a.id, a.name, a.description, a.icon, a.xp_reward, a.condition_type, a.condition_value,
			   ua.earned_at IS NOT NULL as earned, ua.earned_at
		FROM achievements a
		LEFT JOIN user_achievements ua ON a.id = ua.achievement_id AND ua.user_id = ?
		ORDER BY a.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []models.AchievementWithStatus
	for rows.Next() {
		var a models.AchievementWithStatus
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Icon, &a.XPReward,
			&a.ConditionType, &a.ConditionValue, &a.Earned, &a.EarnedAt); err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}
	return achievements, rows.Err()
}

// AwardAchievement grants an achievement to a user
func AwardAchievement(userID, achievementID int64) error {
	_, err := db.DB.Exec(`
		INSERT OR IGNORE INTO user_achievements (user_id, achievement_id)
		VALUES (?, ?)
	`, userID, achievementID)
	return err
}

// HasAchievement checks if user has earned an achievement
func HasAchievement(userID, achievementID int64) (bool, error) {
	var exists bool
	err := db.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM user_achievements WHERE user_id = ? AND achievement_id = ?)
	`, userID, achievementID).Scan(&exists)
	return exists, err
}

// CountEarnedAchievements counts achievements earned by user
func CountEarnedAchievements(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM user_achievements WHERE user_id = ?
	`, userID).Scan(&count)
	return count, err
}

// GetTotalReviews gets total review count for a user
func GetTotalReviews(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM review_logs WHERE user_id = ?
	`, userID).Scan(&count)
	return count, err
}

// GetTotalCardsLearned gets count of cards that have been reviewed at least once
func GetTotalCardsLearned(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(DISTINCT card_id) FROM card_progress
		WHERE user_id = ? AND reps > 0
	`, userID).Scan(&count)
	return count, err
}

// CountTotalReviews counts all reviews by a user
func CountTotalReviews(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM review_logs WHERE user_id = ?
	`, userID).Scan(&count)
	return count, err
}

// CountWordsLearned counts words learned (reviewed at least once with Good or Easy)
func CountWordsLearned(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(DISTINCT card_id) FROM card_progress
		WHERE user_id = ? AND state IN ('learning', 'review') AND reps > 0
	`, userID).Scan(&count)
	return count, err
}
