package models

import (
	"database/sql"
	"time"
)

// User represents a learner
type User struct {
	ID             int64
	Username       string
	CreatedAt      time.Time
	TotalXP        int
	CurrentStreak  int
	LongestStreak  int
	LastActiveDate sql.NullTime
}

// Island represents a vocabulary category/script
type Island struct {
	ID          int64
	Name        string
	Description string
	Icon        string
	UnlockXP    int
	SortOrder   int
}

// Card represents a vocabulary item
type Card struct {
	ID              int64
	IslandID        sql.NullInt64
	Term            string
	Translation     string
	ExampleSentence string
	Notes           string
	AudioURL        string
	FrequencyRank   sql.NullInt64
	CreatedAt       time.Time
	// Joined data
	Bridges []Bridge
}

// Bridge represents a polyglot connection (Hindi/Dutch/English)
type Bridge struct {
	ID           int64
	CardID       int64
	BridgeType   BridgeType
	BridgeContent string
	Explanation  string
}

// BridgeType enum
type BridgeType string

const (
	BridgeHindiPhonetic BridgeType = "hindi_phonetic"
	BridgeDutchSyntax   BridgeType = "dutch_syntax"
	BridgeEnglishCognate BridgeType = "english_cognate"
)

// CardProgress represents FSRS scheduling data for a user's card
type CardProgress struct {
	ID            int64
	UserID        int64
	CardID        int64
	Stability     float64
	Difficulty    float64
	ElapsedDays   int
	ScheduledDays int
	Reps          int
	Lapses        int
	State         CardState
	Due           sql.NullTime
	LastReview    sql.NullTime
}

// CardState enum for FSRS
type CardState string

const (
	StateNew        CardState = "new"
	StateLearning   CardState = "learning"
	StateReview     CardState = "review"
	StateRelearning CardState = "relearning"
)

// Rating enum for reviews (FSRS compatible)
type Rating int

const (
	RatingAgain Rating = 1
	RatingHard  Rating = 2
	RatingGood  Rating = 3
	RatingEasy  Rating = 4
)

// ReviewLog represents a single review event
type ReviewLog struct {
	ID              int64
	UserID          int64
	CardID          int64
	Rating          Rating
	ElapsedDays     int
	ScheduledDays   int
	ReviewedAt      time.Time
	ReviewDurationMs int
}

// DailyLog represents daily activity stats (for heat map)
type DailyLog struct {
	ID            int64
	UserID        int64
	Date          time.Time
	XPEarned      int
	CardsReviewed int
	CardsCorrect  int
	MinutesActive int
	NewCardsAdded int
}

// Achievement represents a gamification badge
type Achievement struct {
	ID             int64
	Name           string
	Description    string
	Icon           string
	XPReward       int
	ConditionType  string
	ConditionValue int
}

// UserAchievement represents an earned badge
type UserAchievement struct {
	UserID        int64
	AchievementID int64
	EarnedAt      time.Time
}

// CardWithProgress combines a card with its FSRS progress
type CardWithProgress struct {
	Card
	Progress *CardProgress
}

// IslandStats represents progress stats for an island
type IslandStats struct {
	Island
	TotalCards    int
	LearnedCards  int
	DueCards      int
	MasteredCards int
}

// ReviewSession represents an active review session
type ReviewSession struct {
	UserID       int64
	Cards        []CardWithProgress
	CurrentIndex int
	StartedAt    time.Time
	XPEarned     int
	Reviewed     int
	Correct      int
}

// StreakInfo contains streak-related data
type StreakInfo struct {
	CurrentStreak int
	LongestStreak int
	LastActive    sql.NullTime
	IsActiveToday bool
}

// TodayStats contains today's activity summary
type TodayStats struct {
	XPEarned      int
	CardsReviewed int
	CardsCorrect  int
	DueCount      int
	NewCount      int
}
