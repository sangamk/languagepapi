package service

import (
	"fmt"
	"math/rand"
	"time"

	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// CurriculumPhases defines the 14-day intensive sprint (~70 new words/day)
var CurriculumPhases = []models.CurriculumPhase{
	{
		ID:             1,
		Name:           "Sprint Week 1",
		Description:    "Core 500 most frequent words",
		StartDay:       1,
		EndDay:         7,
		NewCardsPerDay: 72,
		TargetIslands:  []int64{1, 2, 3}, // Core Essentials, Common Words, Expanding
		ModeWeights:    models.ModeWeights{Standard: 60, Reverse: 30, Typing: 10},
	},
	{
		ID:             2,
		Name:           "Sprint Week 2",
		Description:    "Advanced 500 words + Review",
		StartDay:       8,
		EndDay:         14,
		NewCardsPerDay: 72,
		TargetIslands:  []int64{4, 5, 6, 7, 8, 9}, // All remaining islands
		ModeWeights:    models.ModeWeights{Standard: 50, Reverse: 35, Typing: 15},
	},
}

// LessonService handles the daily lesson flow
type LessonService struct {
	reviewService *ReviewService
}

// NewLessonService creates a new lesson service
func NewLessonService() *LessonService {
	return &LessonService{
		reviewService: NewReviewService(),
	}
}

// GetPhaseForDay returns the curriculum phase for a given day number
func GetPhaseForDay(day int) *models.CurriculumPhase {
	for i := range CurriculumPhases {
		if day >= CurriculumPhases[i].StartDay && day <= CurriculumPhases[i].EndDay {
			return &CurriculumPhases[i]
		}
	}
	// Default to last phase if beyond 14 days
	return &CurriculumPhases[len(CurriculumPhases)-1]
}

// CalculateDayNumber calculates which day of the journey we're on
func CalculateDayNumber(startDate time.Time) int {
	now := time.Now()
	// Normalize to start of day
	startDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	days := int(today.Sub(startDay).Hours() / 24) + 1
	if days < 1 {
		days = 1
	}
	if days > 14 {
		days = 14
	}
	return days
}

// BuildDailyLesson creates today's lesson with smart card selection and mode assignment
func (s *LessonService) BuildDailyLesson(userID int64) (*models.DailyLesson, error) {
	// Get or create journey
	journey, err := repository.GetOrCreateJourney(userID)
	if err != nil {
		return nil, err
	}

	dayNumber := CalculateDayNumber(journey.StartDate)
	phase := GetPhaseForDay(dayNumber)

	// Get all due reviews (mandatory) - curriculum cards only
	dueCards, err := repository.GetDueCards(userID, 100)
	if err != nil {
		return nil, err
	}

	// Get due song vocab cards (max 5 per day)
	songVocabDue, err := repository.GetDueSongVocabCards(userID, 5)
	if err != nil {
		songVocabDue = nil // Continue without song vocab if error
	}

	// Get new song vocab cards from songs in progress (max 2 per day)
	songsInProgress, _ := repository.GetUserSongsInProgress(userID)
	songVocabNew, err := repository.GetNewSongVocabCards(userID, songsInProgress, 2)
	if err != nil {
		songVocabNew = nil
	}

	// Calculate how many new cards based on workload
	newCardsToAdd := s.calculateNewCardsForToday(phase, len(dueCards))

	// Get new cards from target islands
	newCards, err := repository.GetNewCardsFromIslands(userID, phase.TargetIslands, newCardsToAdd)
	if err != nil {
		return nil, err
	}

	// Merge song vocab into due and new card pools
	allDueCards := dueCards
	for _, c := range songVocabDue {
		allDueCards = append(allDueCards, c)
	}
	allNewCards := newCards
	for _, c := range songVocabNew {
		allNewCards = append(allNewCards, c)
	}

	// Build lesson with interleaving
	lessonCards := s.interleaveLessonCards(allDueCards, allNewCards, phase)

	// Mark song vocab cards and fetch song titles
	for i := range lessonCards {
		if lessonCards[i].Source == "song" {
			lessonCards[i].IsSongVocab = true
			if title, err := repository.GetSongTitleForCard(lessonCards[i].ID); err == nil {
				lessonCards[i].SongTitle = title
			}
		}
	}

	// Assign modes to each card
	modeHistory := []string{}
	for i := range lessonCards {
		mode := s.selectModeForCard(&lessonCards[i], phase, modeHistory)
		lessonCards[i].Mode = mode
		modeHistory = append(modeHistory, mode)
	}

	// Estimate duration (~1.5 min per card avg)
	estimatedMins := (len(lessonCards) * 3) / 2
	if estimatedMins < 1 && len(lessonCards) > 0 {
		estimatedMins = 1
	}

	return &models.DailyLesson{
		DayNumber:      dayNumber,
		Phase:          phase,
		Cards:          lessonCards,
		EstimatedMins:  estimatedMins,
		DueReviewCount: len(dueCards) + len(songVocabDue),
		NewCardCount:   len(newCards) + len(songVocabNew),
	}, nil
}

// calculateNewCardsForToday adjusts new card count based on review load
// For intensive sprint: stay aggressive, only minor reductions for extreme loads
func (s *LessonService) calculateNewCardsForToday(phase *models.CurriculumPhase, dueCount int) int {
	base := phase.NewCardsPerDay

	// Only reduce new cards if review load is extreme
	if dueCount > 150 {
		return max(50, base-20) // Still learn 50+ new words
	}
	if dueCount > 100 {
		return max(60, base-10) // Still learn 60+ new words
	}

	return base
}

// interleaveLessonCards mixes reviews and new cards
func (s *LessonService) interleaveLessonCards(
	reviews []models.CardWithProgress,
	newCards []models.CardWithProgress,
	phase *models.CurriculumPhase,
) []models.LessonCard {
	result := make([]models.LessonCard, 0, len(reviews)+len(newCards))

	if len(newCards) == 0 {
		// All reviews
		for _, c := range reviews {
			result = append(result, models.LessonCard{CardWithProgress: c, IsNew: false})
		}
		return result
	}

	// Distribute new cards evenly among reviews
	// Pattern: R R R N R R R N ... (reviews between new cards)
	interval := 3
	if len(reviews) > 0 && len(newCards) > 0 {
		interval = (len(reviews) + 1) / (len(newCards) + 1)
		if interval < 3 {
			interval = 3 // At least 3 reviews between new cards
		}
	}

	reviewIdx, newIdx := 0, 0
	for reviewIdx < len(reviews) || newIdx < len(newCards) {
		// Add reviews
		for i := 0; i < interval && reviewIdx < len(reviews); i++ {
			result = append(result, models.LessonCard{
				CardWithProgress: reviews[reviewIdx],
				IsNew:            false,
			})
			reviewIdx++
		}
		// Add one new card
		if newIdx < len(newCards) {
			result = append(result, models.LessonCard{
				CardWithProgress: newCards[newIdx],
				IsNew:            true,
			})
			newIdx++
		}
	}

	return result
}

// selectModeForCard determines the practice mode based on card state and phase
func (s *LessonService) selectModeForCard(
	card *models.LessonCard,
	phase *models.CurriculumPhase,
	modeHistory []string,
) string {
	// Rule 1: New cards can be standard or mcq (50/50)
	if card.Progress == nil || card.Progress.State == models.StateNew {
		if rand.Intn(100) < 50 {
			return "mcq"
		}
		return "standard"
	}

	// Rule 2: Cards in relearning (lapses > 2) prefer Standard or MCQ
	if card.Progress.State == models.StateRelearning || card.Progress.Lapses > 2 {
		if rand.Intn(100) < 30 {
			return "mcq"
		}
		return s.weightedRandomWithBias("standard", 70, phase.ModeWeights)
	}

	// Rule 3: Low stability cards (< 5 days) - standard, reverse, or mcq
	if card.Progress.Stability < 5 {
		if rand.Intn(100) < 25 {
			return "mcq"
		}
		return s.weightedRandomExcluding("typing", phase.ModeWeights)
	}

	// Rule 4: Medium stability (5-21 days) - can include fill_blank
	if card.Progress.Stability >= 5 && card.Progress.Stability <= 21 {
		r := rand.Intn(100)
		if r < 20 {
			return "mcq"
		}
		if r < 35 {
			return "fill_blank"
		}
		return s.weightedRandom(phase.ModeWeights)
	}

	// Rule 5: High stability cards (> 21 days, mature) - all modes including sentence_build
	if card.Progress.Stability > 21 && card.Progress.Reps >= 5 {
		r := rand.Intn(100)
		if r < 15 {
			return "mcq"
		}
		if r < 30 {
			return "fill_blank"
		}
		if r < 45 {
			return "sentence_build"
		}
		return s.weightedRandomWithBias("typing", 40, phase.ModeWeights)
	}

	// Rule 6: Avoid 3+ of same mode in a row
	if len(modeHistory) >= 2 {
		lastTwo := modeHistory[len(modeHistory)-2:]
		if lastTwo[0] == lastTwo[1] {
			return s.weightedRandomExcluding(lastTwo[0], phase.ModeWeights)
		}
	}

	// Default: weighted random based on phase
	return s.weightedRandom(phase.ModeWeights)
}

// weightedRandom picks a mode based on phase weights
func (s *LessonService) weightedRandom(weights models.ModeWeights) string {
	total := weights.Standard + weights.Reverse + weights.Typing
	r := rand.Intn(total)
	if r < weights.Standard {
		return "standard"
	}
	if r < weights.Standard+weights.Reverse {
		return "reverse"
	}
	return "typing"
}

// weightedRandomWithBias biases toward a specific mode
func (s *LessonService) weightedRandomWithBias(preferred string, biasPercent int, weights models.ModeWeights) string {
	if rand.Intn(100) < biasPercent {
		return preferred
	}
	return s.weightedRandom(weights)
}

// weightedRandomExcluding picks a mode excluding one option
func (s *LessonService) weightedRandomExcluding(exclude string, weights models.ModeWeights) string {
	var standard, reverse, typing int
	switch exclude {
	case "standard":
		reverse = weights.Reverse
		typing = weights.Typing
	case "reverse":
		standard = weights.Standard
		typing = weights.Typing
	case "typing":
		standard = weights.Standard
		reverse = weights.Reverse
	}

	total := standard + reverse + typing
	if total == 0 {
		return "standard" // Fallback
	}

	r := rand.Intn(total)
	if r < standard {
		return "standard"
	}
	if r < standard+reverse {
		return "reverse"
	}
	return "typing"
}

// GetJourneyHomeData builds the data for the journey home page
func (s *LessonService) GetJourneyHomeData(userID int64) (*models.JourneyHomeData, error) {
	// Get or create journey
	journey, err := repository.GetOrCreateJourney(userID)
	if err != nil {
		return nil, err
	}

	dayNumber := CalculateDayNumber(journey.StartDate)
	phase := GetPhaseForDay(dayNumber)

	// Check if today's lesson is complete
	completed, session, err := repository.IsTodayLessonComplete(userID)
	if err != nil {
		return nil, err
	}

	todayStats := ""
	if completed && session != nil {
		accuracy := 0
		if session.CardsReviewed > 0 {
			accuracy = session.CardsCorrect * 100 / session.CardsReviewed
		}
		todayStats = formatStats(session.CardsReviewed, accuracy)
	}

	// Get counts for lesson preview
	dueCount, _ := repository.CountDueCards(userID)

	// Preview new cards
	lesson, err := s.BuildDailyLesson(userID)
	newCount := 0
	estimatedMins := 0
	if err == nil && lesson != nil {
		newCount = lesson.NewCardCount
		estimatedMins = lesson.EstimatedMins
	}

	// Get user for streak and XP
	user, _ := repository.GetUser(userID)
	streak := 0
	totalXP := 0
	if user != nil {
		streak = user.CurrentStreak
		totalXP = user.TotalXP
	}

	return &models.JourneyHomeData{
		DayNumber:      dayNumber,
		TotalDays:      14,
		CurrentPhase:   phase,
		TodayCompleted: completed,
		TodayStats:     todayStats,
		DueCount:       dueCount,
		NewCount:       newCount,
		EstimatedMins:  estimatedMins,
		Streak:         streak,
		TotalXP:        totalXP,
	}, nil
}

func formatStats(reviewed int, accuracy int) string {
	if reviewed == 0 {
		return "No cards reviewed"
	}
	return fmt.Sprintf("%d cards, %d%% accuracy", reviewed, accuracy)
}

// GetMotivationalMessage returns a message based on performance
func GetMotivationalMessage(dayNumber int, accuracy int, isComplete bool) string {
	if !isComplete {
		return "Your daily lesson awaits!"
	}

	messages := []string{
		"Great job today!",
		"You're making progress!",
		"Keep up the momentum!",
		"One step closer to fluency!",
		"Your consistency is paying off!",
	}

	if accuracy >= 90 {
		messages = []string{
			"Outstanding accuracy!",
			"You're crushing it!",
			"Nearly perfect!",
		}
	} else if accuracy >= 80 {
		messages = []string{
			"Solid performance!",
			"Great recall today!",
			"Keep it up!",
		}
	} else if accuracy < 60 {
		messages = []string{
			"Every mistake is a lesson!",
			"You're building foundations!",
			"Practice makes progress!",
		}
	}

	// Milestone messages for 14-day sprint
	switch {
	case dayNumber == 1:
		return "Day 1! Let's crush 1000 words in 2 weeks."
	case dayNumber == 3:
		return "Day 3! You've already learned ~200 words!"
	case dayNumber == 7:
		return "Week 1 complete! ~500 words down, 500 to go!"
	case dayNumber == 10:
		return "Day 10! Over 700 words learned. The finish line is close!"
	case dayNumber == 14:
		return "YOU DID IT! 1000 words in 14 days. Absolute legend."
	}

	return messages[rand.Intn(len(messages))]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
