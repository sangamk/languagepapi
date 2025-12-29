package handlers

import (
	"net/http"

	"languagepapi/components"
	"languagepapi/internal/repository"
)

// HandleHome renders the home page with dynamic stats
func HandleHome(w http.ResponseWriter, r *http.Request) {
	// Get user stats
	user, _ := repository.GetUser(defaultUserID)
	cardCount, _ := repository.CountCards()
	dueCount, _ := repository.CountDueCards(defaultUserID)

	streak := 0
	totalXP := 0
	if user != nil {
		streak = user.CurrentStreak
		totalXP = user.TotalXP
	}

	components.Home(cardCount, streak, dueCount, totalXP).Render(r.Context(), w)
}
