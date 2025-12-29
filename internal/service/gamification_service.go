package service

import (
	"math"
	"time"

	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

const DefaultDailyGoal = 20 // cards per day

// CalculateLevel returns level based on XP (level = floor(sqrt(XP / 100)))
func CalculateLevel(xp int) int {
	if xp <= 0 {
		return 1
	}
	return int(math.Floor(math.Sqrt(float64(xp)/100))) + 1
}

// XPForLevel returns XP required to reach a given level
func XPForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	return (level - 1) * (level - 1) * 100
}

// GetGamificationStats retrieves all gamification data for a user
func GetGamificationStats(userID int64) (*models.GamificationStats, error) {
	user, err := repository.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// Calculate level info
	level := CalculateLevel(user.TotalXP)
	currentLevelXP := XPForLevel(level)
	nextLevelXP := XPForLevel(level + 1)
	xpInCurrentLevel := user.TotalXP - currentLevelXP
	xpNeededForNext := nextLevelXP - currentLevelXP
	xpProgress := float64(xpInCurrentLevel) / float64(xpNeededForNext) * 100

	// Check if active today
	today := time.Now().Format("2006-01-02")
	isActiveToday := false
	if user.LastActiveDate.Valid {
		isActiveToday = user.LastActiveDate.Time.Format("2006-01-02") == today
	}

	// Get today's progress for daily goal
	todayStats, _ := repository.GetTodayStats(userID)
	dailyProgress := 0
	if todayStats != nil {
		dailyProgress = todayStats.CardsReviewed
	}
	dailyPercent := math.Min(float64(dailyProgress)/float64(DefaultDailyGoal)*100, 100)

	// Get achievements
	achievements, _ := repository.GetUserAchievements(userID)
	earnedCount := 0
	for _, a := range achievements {
		if a.Earned {
			earnedCount++
		}
	}

	return &models.GamificationStats{
		Level:          level,
		CurrentXP:      user.TotalXP,
		XPForNextLevel: nextLevelXP,
		XPProgress:     xpProgress,
		CurrentStreak:  user.CurrentStreak,
		LongestStreak:  user.LongestStreak,
		IsActiveToday:  isActiveToday,
		DailyGoal:      DefaultDailyGoal,
		DailyProgress:  dailyProgress,
		DailyPercent:   dailyPercent,
		Achievements:   achievements,
		TotalBadges:    len(achievements),
		EarnedBadges:   earnedCount,
	}, nil
}

// BuildHeatmap creates a 2D grid of days for the heatmap (last N weeks)
func BuildHeatmap(userID int64, weeks int) ([][]models.HeatmapDay, error) {
	days := weeks * 7
	logs, err := repository.GetHeatMapData(userID, days)
	if err != nil {
		return nil, err
	}

	// Create a map of date -> log for quick lookup
	logMap := make(map[string]models.DailyLog)
	for _, log := range logs {
		logMap[log.Date.Format("2006-01-02")] = log
	}

	// Calculate the starting point (go back to first day of that week)
	now := time.Now()
	// Find the Sunday of the current week
	daysUntilSunday := int(now.Weekday())
	endDate := now
	startDate := endDate.AddDate(0, 0, -(days-1)-daysUntilSunday)

	// Build the grid: 7 rows (Sun-Sat) x N weeks
	grid := make([][]models.HeatmapDay, 7)
	for i := range grid {
		grid[i] = make([]models.HeatmapDay, weeks)
	}

	currentDate := startDate
	for week := 0; week < weeks; week++ {
		for day := 0; day < 7; day++ {
			dateStr := currentDate.Format("2006-01-02")
			hd := models.HeatmapDay{
				Date:          dateStr,
				XPEarned:      0,
				CardsReviewed: 0,
				Level:         0,
			}

			if log, ok := logMap[dateStr]; ok {
				hd.XPEarned = log.XPEarned
				hd.CardsReviewed = log.CardsReviewed
				hd.Level = xpToLevel(log.XPEarned)
			}

			// Only show if date is not in the future
			if currentDate.After(now) {
				hd.Level = -1 // Mark as future/empty
			}

			grid[day][week] = hd
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	return grid, nil
}

// xpToLevel converts XP earned in a day to a 0-4 intensity level
func xpToLevel(xp int) int {
	switch {
	case xp == 0:
		return 0
	case xp <= 10:
		return 1
	case xp <= 30:
		return 2
	case xp <= 50:
		return 3
	default:
		return 4
	}
}

// CheckAndAwardAchievements checks conditions and awards any new achievements
func CheckAndAwardAchievements(userID int64) ([]models.Achievement, error) {
	var newlyAwarded []models.Achievement

	user, err := repository.GetUser(userID)
	if err != nil {
		return nil, err
	}

	totalReviews, _ := repository.GetTotalReviews(userID)
	wordsLearned, _ := repository.GetTotalCardsLearned(userID)

	achievements, err := repository.GetUserAchievements(userID)
	if err != nil {
		return nil, err
	}

	for _, a := range achievements {
		if a.Earned {
			continue // Already earned
		}

		earned := false
		switch a.ConditionType {
		case "cards_reviewed":
			earned = totalReviews >= a.ConditionValue
		case "streak":
			earned = user.CurrentStreak >= a.ConditionValue || user.LongestStreak >= a.ConditionValue
		case "words_learned":
			earned = wordsLearned >= a.ConditionValue
		}

		if earned {
			if err := repository.AwardAchievement(userID, a.ID); err == nil {
				newlyAwarded = append(newlyAwarded, a.Achievement)
				// Also award XP for achievement
				repository.UpdateUserXP(userID, a.XPReward)
			}
		}
	}

	return newlyAwarded, nil
}

// GetCalendarData retrieves all data needed for the calendar page
func GetCalendarData(userID int64) (*models.CalendarData, error) {
	stats, err := GetGamificationStats(userID)
	if err != nil {
		return nil, err
	}

	heatmap, err := BuildHeatmap(userID, 20) // ~5 months
	if err != nil {
		return nil, err
	}

	totalReviews, _ := repository.GetTotalReviews(userID)
	totalCards, _ := repository.CountCards()
	user, _ := repository.GetUser(userID)

	totalXP := 0
	if user != nil {
		totalXP = user.TotalXP
	}

	return &models.CalendarData{
		Heatmap:      heatmap,
		Stats:        *stats,
		TotalXP:      totalXP,
		TotalReviews: totalReviews,
		TotalCards:   totalCards,
	}, nil
}
