package service

import (
	"database/sql"
	"time"

	"languagepapi/internal/fsrs"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// ReviewService handles the review flow
type ReviewService struct {
	fsrs *fsrs.Service
}

// NewReviewService creates a new review service
func NewReviewService() *ReviewService {
	return &ReviewService{
		fsrs: fsrs.NewService(),
	}
}

// StartSession creates a new review session with due and new cards
func (s *ReviewService) StartSession(userID int64, maxCards int) (*models.ReviewSession, error) {
	session := &models.ReviewSession{
		UserID:    userID,
		StartedAt: time.Now(),
	}

	// Get due cards first (priority)
	dueCards, err := repository.GetDueCards(userID, maxCards)
	if err != nil {
		return nil, err
	}
	session.Cards = append(session.Cards, dueCards...)

	// Fill remaining slots with new cards
	remaining := maxCards - len(dueCards)
	if remaining > 0 {
		newCards, err := repository.GetNewCards(userID, remaining)
		if err != nil {
			return nil, err
		}
		session.Cards = append(session.Cards, newCards...)
	}

	return session, nil
}

// GetNextCard returns the next card to review in the session
func (s *ReviewService) GetNextCard(session *models.ReviewSession) (*models.CardWithProgress, bool) {
	if session.CurrentIndex >= len(session.Cards) {
		return nil, false // Session complete
	}
	card := &session.Cards[session.CurrentIndex]
	return card, true
}

// GetSchedulingPreview returns preview for all rating options
func (s *ReviewService) GetSchedulingPreview(card *models.CardWithProgress) map[models.Rating]fsrs.SchedulingPreview {
	return s.fsrs.GetSchedulingPreview(card.Progress, time.Now())
}

// SubmitReview processes a review and updates the session
func (s *ReviewService) SubmitReview(
	session *models.ReviewSession,
	cardID int64,
	rating models.Rating,
	durationMs int,
) (*ReviewResult, error) {
	now := time.Now()
	userID := session.UserID

	// Find the card in session
	var card *models.CardWithProgress
	for i := range session.Cards {
		if session.Cards[i].ID == cardID {
			card = &session.Cards[i]
			break
		}
	}
	if card == nil {
		return nil, nil // Card not found in session
	}

	// Determine if this is a new card
	isNew := card.Progress == nil || card.Progress.State == models.StateNew

	// Initialize progress if needed
	if card.Progress == nil {
		card.Progress = &models.CardProgress{
			UserID: userID,
			CardID: cardID,
			State:  models.StateNew,
			Due:    sql.NullTime{Time: now, Valid: true},
		}
	}

	// Calculate new schedule using FSRS
	newProgress := s.fsrs.ScheduleReview(card.Progress, rating, now)

	// Save progress
	if err := repository.UpsertProgress(newProgress); err != nil {
		return nil, err
	}

	// Log the review
	reviewLog := &models.ReviewLog{
		UserID:           userID,
		CardID:           cardID,
		Rating:           rating,
		ElapsedDays:      card.Progress.ElapsedDays,
		ScheduledDays:    newProgress.ScheduledDays,
		ReviewDurationMs: durationMs,
	}
	if err := repository.LogReview(reviewLog); err != nil {
		return nil, err
	}

	// Calculate XP
	xp := CalculateReviewXP(rating, isNew, 0) // TODO: pass actual streak

	// Update user XP
	if err := repository.UpdateUserXP(userID, xp); err != nil {
		return nil, err
	}

	// Update streak
	if err := repository.UpdateStreak(userID); err != nil {
		return nil, err
	}

	// Update daily stats
	correct := 0
	if rating >= models.RatingGood {
		correct = 1
	}
	newCardCount := 0
	if isNew {
		newCardCount = 1
	}
	if err := repository.IncrementDailyStats(userID, xp, 1, correct, newCardCount); err != nil {
		return nil, err
	}

	// Update session counters
	session.Reviewed++
	session.XPEarned += xp
	if rating >= models.RatingGood {
		session.Correct++
	}
	session.CurrentIndex++

	// Update the card in session with new progress
	card.Progress = newProgress

	return &ReviewResult{
		XPEarned:      xp,
		NextDue:       newProgress.Due.Time,
		Interval:      newProgress.ScheduledDays,
		SessionDone:   session.CurrentIndex >= len(session.Cards),
		TotalReviewed: session.Reviewed,
		TotalCorrect:  session.Correct,
		TotalXP:       session.XPEarned,
	}, nil
}

// ReviewResult contains the result of a review submission
type ReviewResult struct {
	XPEarned      int
	NextDue       time.Time
	Interval      int
	SessionDone   bool
	TotalReviewed int
	TotalCorrect  int
	TotalXP       int
}

// CalculateReviewXP calculates XP for a review
func CalculateReviewXP(rating models.Rating, isNew bool, streak int) int {
	xp := 2 // Base XP for any review

	// Bonus for correct answers
	if rating >= models.RatingGood {
		xp += 1
	}

	// Bonus for new card learned
	if isNew && rating >= models.RatingGood {
		xp += 3 // Total 6 XP for successfully learning a new card
	}

	// Streak bonus (1 XP per 7-day streak, capped at 5)
	streakBonus := streak / 7
	if streakBonus > 5 {
		streakBonus = 5
	}
	xp += streakBonus

	return xp
}

// GetSessionStats returns current session statistics
func (s *ReviewService) GetSessionStats(session *models.ReviewSession) SessionStats {
	remaining := len(session.Cards) - session.CurrentIndex
	if remaining < 0 {
		remaining = 0
	}

	accuracy := 0.0
	if session.Reviewed > 0 {
		accuracy = float64(session.Correct) / float64(session.Reviewed) * 100
	}

	return SessionStats{
		TotalCards: len(session.Cards),
		Reviewed:   session.Reviewed,
		Remaining:  remaining,
		Correct:    session.Correct,
		Accuracy:   accuracy,
		XPEarned:   session.XPEarned,
		Duration:   time.Since(session.StartedAt),
	}
}

// SessionStats contains session statistics
type SessionStats struct {
	TotalCards int
	Reviewed   int
	Remaining  int
	Correct    int
	Accuracy   float64
	XPEarned   int
	Duration   time.Duration
}

// CheckAchievements checks and awards any new achievements
func (s *ReviewService) CheckAchievements(userID int64) []models.Achievement {
	var newAchievements []models.Achievement

	// Get user stats
	user, err := repository.GetUser(userID)
	if err != nil {
		return newAchievements
	}

	// Get all achievements
	achievements, err := repository.GetAllAchievements()
	if err != nil {
		return newAchievements
	}

	// Get earned achievements
	earned, err := repository.GetUserAchievements(userID)
	if err != nil {
		return newAchievements
	}

	earnedMap := make(map[int64]bool)
	for _, e := range earned {
		earnedMap[e.ID] = true
	}

	// Get total review count
	totalReviews, _ := repository.CountTotalReviews(userID)

	// Get words learned count
	wordsLearned, _ := repository.CountWordsLearned(userID)

	// Check each achievement
	for _, a := range achievements {
		if earnedMap[a.ID] {
			continue // Already earned
		}

		var earned bool
		switch a.ConditionType {
		case "cards_reviewed":
			earned = totalReviews >= a.ConditionValue
		case "streak":
			earned = user.CurrentStreak >= a.ConditionValue
		case "words_learned":
			earned = wordsLearned >= a.ConditionValue
		}

		if earned {
			if err := repository.AwardAchievement(userID, a.ID); err == nil {
				// Add XP reward
				repository.UpdateUserXP(userID, a.XPReward)
				newAchievements = append(newAchievements, a)
			}
		}
	}

	return newAchievements
}
