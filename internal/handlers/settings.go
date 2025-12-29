package handlers

import (
	"net/http"
	"strconv"

	"languagepapi/components"
	"languagepapi/internal/repository"
)

// HandleSettings renders the settings page
func HandleSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := repository.GetUserSettings(defaultUserID)
	if err != nil {
		// Use defaults
		settings = &repository.UserSettings{
			DailyGoal:        20,
			EnableTTS:        true,
			ShowBridges:      true,
			DefaultMode:      "standard",
			NewCardsPerDay:   10,
			ReviewsPerSession: 20,
		}
	}

	components.Settings(settings, "", false).Render(r.Context(), w)
}

// HandleSaveSettings saves user settings
func HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	dailyGoal, _ := strconv.Atoi(r.FormValue("daily_goal"))
	newCardsPerDay, _ := strconv.Atoi(r.FormValue("new_cards_per_day"))
	reviewsPerSession, _ := strconv.Atoi(r.FormValue("reviews_per_session"))
	enableTTS := r.FormValue("enable_tts") == "on"
	showBridges := r.FormValue("show_bridges") == "on"
	defaultMode := r.FormValue("default_mode")

	if dailyGoal < 1 {
		dailyGoal = 20
	}
	if newCardsPerDay < 1 {
		newCardsPerDay = 10
	}
	if reviewsPerSession < 1 {
		reviewsPerSession = 20
	}
	if defaultMode == "" {
		defaultMode = "standard"
	}

	settings := &repository.UserSettings{
		DailyGoal:        dailyGoal,
		EnableTTS:        enableTTS,
		ShowBridges:      showBridges,
		DefaultMode:      defaultMode,
		NewCardsPerDay:   newCardsPerDay,
		ReviewsPerSession: reviewsPerSession,
	}

	if err := repository.SaveUserSettings(defaultUserID, settings); err != nil {
		components.Settings(settings, "Failed to save settings", false).Render(r.Context(), w)
		return
	}

	components.Settings(settings, "Settings saved!", true).Render(r.Context(), w)
}
