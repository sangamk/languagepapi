package service

import (
	"testing"

	"languagepapi/internal/models"
)

func TestCalculateReviewXP(t *testing.T) {
	tests := []struct {
		name     string
		rating   models.Rating
		isNew    bool
		streak   int
		expected int
	}{
		{
			name:     "Again rating, existing card",
			rating:   models.RatingAgain,
			isNew:    false,
			streak:   0,
			expected: 2, // Base XP only
		},
		{
			name:     "Good rating, existing card",
			rating:   models.RatingGood,
			isNew:    false,
			streak:   0,
			expected: 3, // Base XP + correct bonus
		},
		{
			name:     "Good rating, new card",
			rating:   models.RatingGood,
			isNew:    true,
			streak:   0,
			expected: 6, // Base XP + correct bonus + new card bonus
		},
		{
			name:     "Easy rating, new card",
			rating:   models.RatingEasy,
			isNew:    true,
			streak:   0,
			expected: 6, // Same as Good for new card
		},
		{
			name:     "Good rating, with 7-day streak",
			rating:   models.RatingGood,
			isNew:    false,
			streak:   7,
			expected: 4, // Base + correct + streak bonus
		},
		{
			name:     "Good rating, with 14-day streak",
			rating:   models.RatingGood,
			isNew:    false,
			streak:   14,
			expected: 5, // Base + correct + 2 streak bonus
		},
		{
			name:     "Good rating, with 100-day streak",
			rating:   models.RatingGood,
			isNew:    false,
			streak:   100,
			expected: 8, // Base + correct + 5 streak bonus (capped)
		},
		{
			name:     "Hard rating, existing card",
			rating:   models.RatingHard,
			isNew:    false,
			streak:   0,
			expected: 2, // Base XP only, Hard is not "correct"
		},
		{
			name:     "Again rating, new card",
			rating:   models.RatingAgain,
			isNew:    true,
			streak:   0,
			expected: 2, // No new card bonus for Again
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateReviewXP(tt.rating, tt.isNew, tt.streak)
			if result != tt.expected {
				t.Errorf("CalculateReviewXP(%v, %v, %v) = %v, want %v",
					tt.rating, tt.isNew, tt.streak, result, tt.expected)
			}
		})
	}
}

func TestFormatInterval(t *testing.T) {
	tests := []struct {
		days     int
		expected string
	}{
		{0, "<1m"},
		{1, "1d"},
		{7, "7d"},
		{29, "29d"},
		{30, "1mo"},
		{60, "2mo"},
		{365, "1y"},
		{730, "2y"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// The formatInterval function is in practice.templ
			// This test is a placeholder for when it's extracted
		})
	}
}
