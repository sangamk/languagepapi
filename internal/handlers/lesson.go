package handlers

import (
	"net/http"
	"strconv"
	"sync"

	"languagepapi/components"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
	"languagepapi/internal/service"
)

var (
	// In-memory lesson storage (single user for now)
	lessons     = make(map[int64]*models.DailyLesson)
	lessonIdx   = make(map[int64]int) // Current card index
	lessonStats = make(map[int64]*lessonSessionStats)
	lessonsLock sync.RWMutex
)

type lessonSessionStats struct {
	Reviewed    int
	Correct     int
	NewLearned  int
	XPEarned    int
	TotalTimeMs int64
	CardResults []models.CardResult
}

// HandleLessonStart starts the daily lesson
func HandleLessonStart(w http.ResponseWriter, r *http.Request) {
	lessonsLock.Lock()
	defer lessonsLock.Unlock()

	// Build today's lesson
	lesson, err := lessonService.BuildDailyLesson(defaultUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store lesson
	lessons[defaultUserID] = lesson
	lessonIdx[defaultUserID] = 0
	lessonStats[defaultUserID] = &lessonSessionStats{}

	// Check if there are any cards
	if len(lesson.Cards) == 0 {
		components.LessonEmpty(lesson.DayNumber, lesson.Phase).Render(r.Context(), w)
		return
	}

	// Get first card
	card := &lesson.Cards[0]

	// Fetch bridges
	if len(card.Bridges) == 0 {
		bridges, _ := repository.GetBridgesForCard(card.ID)
		card.Bridges = bridges
	}

	// Get scheduling preview
	preview := reviewService.GetSchedulingPreview(&card.CardWithProgress)

	components.LessonCard(card, preview, 1, len(lesson.Cards), lesson.DayNumber, lesson.Phase).Render(r.Context(), w)
}

// HandleLessonReview processes a review within the lesson
func HandleLessonReview(w http.ResponseWriter, r *http.Request) {
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

	lessonsLock.Lock()
	lesson, exists := lessons[defaultUserID]
	if !exists {
		lessonsLock.Unlock()
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	currentIdx := lessonIdx[defaultUserID]
	stats := lessonStats[defaultUserID]

	// Find the card and get its isNew status
	var currentCard *models.LessonCard
	for i := range lesson.Cards {
		if lesson.Cards[i].ID == cardID {
			currentCard = &lesson.Cards[i]
			break
		}
	}

	// Submit review using the existing review service
	// We need to use the standard session mechanism but without mode selection
	// Create a temporary session for this card
	tempSession := &models.ReviewSession{
		UserID: defaultUserID,
		Cards:  []models.CardWithProgress{currentCard.CardWithProgress},
	}

	result, err := reviewService.SubmitReview(tempSession, cardID, models.Rating(rating), durationMs)
	if err != nil {
		lessonsLock.Unlock()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update lesson stats
	stats.Reviewed++
	stats.XPEarned += result.XPEarned
	stats.TotalTimeMs += int64(durationMs)
	if rating >= 3 {
		stats.Correct++
	}
	if currentCard != nil && currentCard.IsNew && rating >= 3 {
		stats.NewLearned++
	}

	// Track per-card result
	cardResult := models.CardResult{
		CardID:      cardID,
		Term:        currentCard.Term,
		Translation: currentCard.Translation,
		Rating:      models.Rating(rating),
		TimeSpentMs: int64(durationMs),
		Mode:        currentCard.Mode,
		WasCorrect:  rating >= 3,
		IsNew:       currentCard.IsNew,
	}
	stats.CardResults = append(stats.CardResults, cardResult)

	// Move to next card
	lessonIdx[defaultUserID] = currentIdx + 1
	nextIdx := currentIdx + 1
	lessonsLock.Unlock()

	// Check if lesson is complete
	if nextIdx >= len(lesson.Cards) {
		// Mark lesson as complete
		_, session, _ := repository.IsTodayLessonComplete(defaultUserID)
		if session == nil {
			repository.CreateLessonSession(defaultUserID, lesson.DayNumber, lesson.Phase.ID)
		}
		session, _ = repository.GetTodayLessonSession(defaultUserID)
		if session != nil {
			repository.UpdateLessonSession(session.ID, stats.Reviewed, stats.Correct, stats.NewLearned, stats.XPEarned)
			repository.CompleteLessonSession(session.ID)
		}

		// Check for new achievements
		newAchievements := reviewService.CheckAchievements(defaultUserID)

		// Calculate accuracy
		accuracy := 0
		if stats.Reviewed > 0 {
			accuracy = stats.Correct * 100 / stats.Reviewed
		}

		// Get motivational message
		message := service.GetMotivationalMessage(lesson.DayNumber, accuracy, true)

		// Build struggles list (cards with rating 1-2)
		var struggles []models.CardResult
		for _, cr := range stats.CardResults {
			if !cr.WasCorrect {
				struggles = append(struggles, cr)
			}
		}

		// Calculate average time per card
		avgTime := int64(0)
		if stats.Reviewed > 0 {
			avgTime = stats.TotalTimeMs / int64(stats.Reviewed)
		}

		// Build summary
		summary := &models.LessonSummary{
			DayNumber:      lesson.DayNumber,
			Phase:          lesson.Phase,
			TotalCards:     stats.Reviewed,
			CorrectCount:   stats.Correct,
			Accuracy:       accuracy,
			TotalTimeMs:    stats.TotalTimeMs,
			AvgTimePerCard: avgTime,
			XPEarned:       stats.XPEarned,
			NewLearned:     stats.NewLearned,
			CardResults:    stats.CardResults,
			Struggles:      struggles,
			Achievements:   newAchievements,
			Message:        message,
		}

		components.LessonCompleteWithSummary(summary).Render(r.Context(), w)
		return
	}

	// Get next card
	lessonsLock.RLock()
	nextCard := &lesson.Cards[nextIdx]
	lessonsLock.RUnlock()

	// Fetch bridges
	if len(nextCard.Bridges) == 0 {
		bridges, _ := repository.GetBridgesForCard(nextCard.ID)
		nextCard.Bridges = bridges
	}

	preview := reviewService.GetSchedulingPreview(&nextCard.CardWithProgress)
	components.LessonCard(nextCard, preview, nextIdx+1, len(lesson.Cards), lesson.DayNumber, lesson.Phase).Render(r.Context(), w)
}

// HandleLessonSkip skips the current card in the lesson
func HandleLessonSkip(w http.ResponseWriter, r *http.Request) {
	lessonsLock.Lock()
	lesson, exists := lessons[defaultUserID]
	if !exists {
		lessonsLock.Unlock()
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	currentIdx := lessonIdx[defaultUserID]
	stats := lessonStats[defaultUserID]
	lessonIdx[defaultUserID] = currentIdx + 1
	nextIdx := currentIdx + 1
	lessonsLock.Unlock()

	// Check if lesson is complete
	if nextIdx >= len(lesson.Cards) {
		// Mark lesson as complete even with skips
		_, session, _ := repository.IsTodayLessonComplete(defaultUserID)
		if session == nil {
			repository.CreateLessonSession(defaultUserID, lesson.DayNumber, lesson.Phase.ID)
		}
		session, _ = repository.GetTodayLessonSession(defaultUserID)
		if session != nil {
			repository.UpdateLessonSession(session.ID, stats.Reviewed, stats.Correct, stats.NewLearned, stats.XPEarned)
			repository.CompleteLessonSession(session.ID)
		}

		newAchievements := reviewService.CheckAchievements(defaultUserID)

		accuracy := 0
		if stats.Reviewed > 0 {
			accuracy = stats.Correct * 100 / stats.Reviewed
		}

		message := service.GetMotivationalMessage(lesson.DayNumber, accuracy, true)

		// Build struggles list
		var struggles []models.CardResult
		for _, cr := range stats.CardResults {
			if !cr.WasCorrect {
				struggles = append(struggles, cr)
			}
		}

		avgTime := int64(0)
		if stats.Reviewed > 0 {
			avgTime = stats.TotalTimeMs / int64(stats.Reviewed)
		}

		summary := &models.LessonSummary{
			DayNumber:      lesson.DayNumber,
			Phase:          lesson.Phase,
			TotalCards:     stats.Reviewed,
			CorrectCount:   stats.Correct,
			Accuracy:       accuracy,
			TotalTimeMs:    stats.TotalTimeMs,
			AvgTimePerCard: avgTime,
			XPEarned:       stats.XPEarned,
			NewLearned:     stats.NewLearned,
			CardResults:    stats.CardResults,
			Struggles:      struggles,
			Achievements:   newAchievements,
			Message:        message,
		}

		components.LessonCompleteWithSummary(summary).Render(r.Context(), w)
		return
	}

	// Get next card
	lessonsLock.RLock()
	nextCard := &lesson.Cards[nextIdx]
	lessonsLock.RUnlock()

	// Fetch bridges
	if len(nextCard.Bridges) == 0 {
		bridges, _ := repository.GetBridgesForCard(nextCard.ID)
		nextCard.Bridges = bridges
	}

	preview := reviewService.GetSchedulingPreview(&nextCard.CardWithProgress)
	components.LessonCard(nextCard, preview, nextIdx+1, len(lesson.Cards), lesson.DayNumber, lesson.Phase).Render(r.Context(), w)
}
