package handlers

import (
	"net/http"
	"strconv"
	"sync"

	"languagepapi/components"
	"languagepapi/internal/models"
	"languagepapi/internal/service"
)

var (
	reviewService = service.NewReviewService()
	// In-memory session storage (single user for now)
	sessions     = make(map[int64]*models.ReviewSession)
	sessionsLock sync.RWMutex
)

const defaultUserID int64 = 1

// HandlePractice starts or continues a review session
func HandlePractice(w http.ResponseWriter, r *http.Request) {
	sessionsLock.Lock()
	session, exists := sessions[defaultUserID]
	if !exists || session.CurrentIndex >= len(session.Cards) {
		// Start new session
		var err error
		session, err = reviewService.StartSession(defaultUserID, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			sessionsLock.Unlock()
			return
		}
		sessions[defaultUserID] = session
	}
	sessionsLock.Unlock()

	// Check if session has no cards (empty state)
	if len(session.Cards) == 0 {
		components.PracticeEmpty().Render(r.Context(), w)
		return
	}

	// Get current card
	card, hasMore := reviewService.GetNextCard(session)
	if !hasMore {
		// Session complete
		stats := reviewService.GetSessionStats(session)
		components.PracticeComplete(stats.Reviewed, stats.Correct, stats.XPEarned).Render(r.Context(), w)
		return
	}

	// Get scheduling preview for rating buttons
	preview := reviewService.GetSchedulingPreview(card)

	components.PracticeCard(card, preview, session.CurrentIndex+1, len(session.Cards)).Render(r.Context(), w)
}

// HandlePracticeCard returns just the card content (HTMX partial)
func HandlePracticeCard(w http.ResponseWriter, r *http.Request) {
	sessionsLock.RLock()
	session, exists := sessions[defaultUserID]
	sessionsLock.RUnlock()

	if !exists {
		http.Redirect(w, r, "/practice", http.StatusSeeOther)
		return
	}

	card, hasMore := reviewService.GetNextCard(session)
	if !hasMore {
		stats := reviewService.GetSessionStats(session)
		components.PracticeComplete(stats.Reviewed, stats.Correct, stats.XPEarned).Render(r.Context(), w)
		return
	}

	preview := reviewService.GetSchedulingPreview(card)
	components.PracticeCard(card, preview, session.CurrentIndex+1, len(session.Cards)).Render(r.Context(), w)
}

// HandleReview processes a review submission
func HandleReview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cardID, err := strconv.ParseInt(r.FormValue("card_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid card_id", http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil || rating < 1 || rating > 4 {
		http.Error(w, "invalid rating", http.StatusBadRequest)
		return
	}

	durationMs, _ := strconv.Atoi(r.FormValue("duration_ms"))

	sessionsLock.Lock()
	session, exists := sessions[defaultUserID]
	if !exists {
		sessionsLock.Unlock()
		http.Error(w, "no active session", http.StatusBadRequest)
		return
	}

	result, err := reviewService.SubmitReview(session, cardID, models.Rating(rating), durationMs)
	sessionsLock.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return next card or completion screen
	if result.SessionDone {
		components.PracticeComplete(result.TotalReviewed, result.TotalCorrect, result.TotalXP).Render(r.Context(), w)
		return
	}

	// Get next card
	sessionsLock.RLock()
	card, _ := reviewService.GetNextCard(session)
	cardCount := len(session.Cards)
	cardIndex := session.CurrentIndex + 1
	sessionsLock.RUnlock()

	preview := reviewService.GetSchedulingPreview(card)
	components.PracticeCard(card, preview, cardIndex, cardCount).Render(r.Context(), w)
}

// HandlePracticeStats returns session stats (HTMX partial)
func HandlePracticeStats(w http.ResponseWriter, r *http.Request) {
	sessionsLock.RLock()
	session, exists := sessions[defaultUserID]
	sessionsLock.RUnlock()

	if !exists {
		http.Error(w, "no active session", http.StatusBadRequest)
		return
	}

	stats := reviewService.GetSessionStats(session)
	components.PracticeStats(stats.Reviewed, stats.Remaining, stats.XPEarned).Render(r.Context(), w)
}
