package handlers

import (
	"net/http"

	"languagepapi/components"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// HandleProgress renders the progress overview page
func HandleProgress(w http.ResponseWriter, r *http.Request) {
	// Get overview stats
	stats, err := repository.GetProgressOverviewStats(defaultUserID)
	if err != nil {
		http.Error(w, "Failed to load stats", http.StatusInternalServerError)
		return
	}

	// Get progress by island
	islands, err := repository.GetLearnedCardsByIsland(defaultUserID)
	if err != nil {
		http.Error(w, "Failed to load island progress", http.StatusInternalServerError)
		return
	}

	// Get recently learned words
	recentWords, err := repository.GetRecentlyLearnedWords(defaultUserID, 20)
	if err != nil {
		http.Error(w, "Failed to load recent words", http.StatusInternalServerError)
		return
	}

	// Get recent lesson sessions
	recentLessons, err := repository.GetRecentLessonSessions(defaultUserID, 7)
	if err != nil {
		http.Error(w, "Failed to load recent lessons", http.StatusInternalServerError)
		return
	}

	// Get user stats for XP and streak
	user, err := repository.GetUser(defaultUserID)
	totalXP := 0
	currentStreak := 0
	if err == nil && user != nil {
		totalXP = user.TotalXP
		currentStreak = user.CurrentStreak
	}

	data := &models.ProgressOverviewData{
		Stats:         stats,
		Islands:       islands,
		RecentWords:   recentWords,
		RecentLessons: recentLessons,
		TotalXP:       totalXP,
		CurrentStreak: currentStreak,
	}

	components.Progress(data).Render(r.Context(), w)
}
