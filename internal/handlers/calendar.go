package handlers

import (
	"net/http"

	"languagepapi/components"
	"languagepapi/internal/service"
)

// HandleCalendar renders the calendar/stats page
func HandleCalendar(w http.ResponseWriter, r *http.Request) {
	data, err := service.GetCalendarData(defaultUserID)
	if err != nil {
		http.Error(w, "Failed to load calendar data", http.StatusInternalServerError)
		return
	}

	// Check and award any new achievements
	service.CheckAndAwardAchievements(defaultUserID)

	components.Calendar(data).Render(r.Context(), w)
}
