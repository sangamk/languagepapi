package fsrs

import (
	"database/sql"
	"time"

	gofsrs "github.com/open-spaced-repetition/go-fsrs/v3"
	"languagepapi/internal/models"
)

// Service wraps the go-fsrs library with app-specific helpers
type Service struct {
	fsrs *gofsrs.FSRS
}

// NewService creates a new FSRS service with default parameters
func NewService() *Service {
	return &Service{
		fsrs: gofsrs.NewFSRS(gofsrs.DefaultParam()),
	}
}

// NewServiceWithRetention creates an FSRS service with custom retention target
func NewServiceWithRetention(requestRetention float64) *Service {
	params := gofsrs.DefaultParam()
	params.RequestRetention = requestRetention
	return &Service{fsrs: gofsrs.NewFSRS(params)}
}

// toFSRSCard converts our CardProgress to go-fsrs Card
func (s *Service) toFSRSCard(p *models.CardProgress) gofsrs.Card {
	card := gofsrs.NewCard()

	if p == nil {
		return card
	}

	card.Stability = p.Stability
	card.Difficulty = p.Difficulty
	card.ElapsedDays = uint64(p.ElapsedDays)
	card.ScheduledDays = uint64(p.ScheduledDays)
	card.Reps = uint64(p.Reps)
	card.Lapses = uint64(p.Lapses)

	switch p.State {
	case models.StateNew:
		card.State = gofsrs.New
	case models.StateLearning:
		card.State = gofsrs.Learning
	case models.StateReview:
		card.State = gofsrs.Review
	case models.StateRelearning:
		card.State = gofsrs.Relearning
	}

	if p.Due.Valid {
		card.Due = p.Due.Time
	}
	if p.LastReview.Valid {
		card.LastReview = p.LastReview.Time
	}

	return card
}

// fromFSRSCard converts go-fsrs Card back to our CardProgress
func (s *Service) fromFSRSCard(card gofsrs.Card, p *models.CardProgress) {
	p.Stability = card.Stability
	p.Difficulty = card.Difficulty
	p.ElapsedDays = int(card.ElapsedDays)
	p.ScheduledDays = int(card.ScheduledDays)
	p.Reps = int(card.Reps)
	p.Lapses = int(card.Lapses)

	switch card.State {
	case gofsrs.New:
		p.State = models.StateNew
	case gofsrs.Learning:
		p.State = models.StateLearning
	case gofsrs.Review:
		p.State = models.StateReview
	case gofsrs.Relearning:
		p.State = models.StateRelearning
	}

	p.Due = sql.NullTime{Time: card.Due, Valid: true}
	p.LastReview = sql.NullTime{Time: card.LastReview, Valid: !card.LastReview.IsZero()}
}

// toFSRSRating converts our Rating to go-fsrs Rating
func toFSRSRating(r models.Rating) gofsrs.Rating {
	switch r {
	case models.RatingAgain:
		return gofsrs.Again
	case models.RatingHard:
		return gofsrs.Hard
	case models.RatingGood:
		return gofsrs.Good
	case models.RatingEasy:
		return gofsrs.Easy
	default:
		return gofsrs.Good
	}
}

// ScheduleReview calculates the next review schedule based on rating
func (s *Service) ScheduleReview(p *models.CardProgress, rating models.Rating, now time.Time) *models.CardProgress {
	card := s.toFSRSCard(p)
	fsrsRating := toFSRSRating(rating)

	// Get scheduling info for this rating
	schedulingInfo := s.fsrs.Repeat(card, now)
	newCard := schedulingInfo[fsrsRating].Card

	// Update progress with new values
	result := &models.CardProgress{
		ID:     p.ID,
		UserID: p.UserID,
		CardID: p.CardID,
	}
	s.fromFSRSCard(newCard, result)

	return result
}

// GetSchedulingPreview returns scheduling info for all possible ratings
func (s *Service) GetSchedulingPreview(p *models.CardProgress, now time.Time) map[models.Rating]SchedulingPreview {
	card := s.toFSRSCard(p)
	schedulingInfo := s.fsrs.Repeat(card, now)

	preview := make(map[models.Rating]SchedulingPreview)

	for rating, fsrsRating := range map[models.Rating]gofsrs.Rating{
		models.RatingAgain: gofsrs.Again,
		models.RatingHard:  gofsrs.Hard,
		models.RatingGood:  gofsrs.Good,
		models.RatingEasy:  gofsrs.Easy,
	} {
		info := schedulingInfo[fsrsRating]
		// Calculate retrievability at the scheduled due date
		retrievability := s.fsrs.GetRetrievability(info.Card, info.Card.Due)
		preview[rating] = SchedulingPreview{
			NextDue:        info.Card.Due,
			Interval:       int(info.Card.ScheduledDays),
			Retrievability: retrievability,
		}
	}

	return preview
}

// SchedulingPreview contains preview info for a rating option
type SchedulingPreview struct {
	NextDue        time.Time
	Interval       int     // days until next review
	Retrievability float64 // estimated memory strength
}

// GetDueDate returns when the card is due
func (s *Service) GetDueDate(p *models.CardProgress) time.Time {
	if p == nil || !p.Due.Valid {
		return time.Now()
	}
	return p.Due.Time
}

// Retrievability calculates current memory strength (0-1)
func (s *Service) Retrievability(p *models.CardProgress, now time.Time) float64 {
	if p == nil || p.Stability == 0 {
		return 0
	}

	card := s.toFSRSCard(p)
	return s.fsrs.GetRetrievability(card, now)
}

// IsDue checks if a card is due for review
func (s *Service) IsDue(p *models.CardProgress, now time.Time) bool {
	if p == nil {
		return true // New cards are always "due"
	}
	if !p.Due.Valid {
		return true
	}
	return now.After(p.Due.Time) || now.Equal(p.Due.Time)
}

// FormatInterval returns a human-readable interval string
func FormatInterval(days int) string {
	if days == 0 {
		return "now"
	}
	if days == 1 {
		return "1d"
	}
	if days < 30 {
		return string(rune('0'+days/10)) + string(rune('0'+days%10)) + "d"
	}
	if days < 365 {
		months := days / 30
		return string(rune('0'+months/10)) + string(rune('0'+months%10)) + "mo"
	}
	years := days / 365
	return string(rune('0'+years)) + "y"
}
