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
	Source          string        // "curriculum" or "song"
	SourceSongID    sql.NullInt64 // Reference to songs.id if source="song"
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

// AchievementWithStatus combines achievement with earned status
type AchievementWithStatus struct {
	Achievement
	Earned   bool
	EarnedAt sql.NullTime
}

// HeatmapDay represents a single day in the heatmap
type HeatmapDay struct {
	Date          string
	XPEarned      int
	CardsReviewed int
	Level         int // 0-4 intensity level
}

// GamificationStats holds all gamification data for display
type GamificationStats struct {
	// Level system
	Level           int
	CurrentXP       int
	XPForNextLevel  int
	XPProgress      float64 // 0-100 percentage

	// Streaks
	CurrentStreak   int
	LongestStreak   int
	IsActiveToday   bool

	// Daily goal
	DailyGoal       int
	DailyProgress   int
	DailyPercent    float64

	// Achievements
	Achievements    []AchievementWithStatus
	TotalBadges     int
	EarnedBadges    int
}

// CalendarData holds all data for the calendar page
type CalendarData struct {
	Heatmap         [][]HeatmapDay // 7 rows (days of week) x N weeks
	Stats           GamificationStats
	TotalXP         int
	TotalReviews    int
	TotalCards      int
}

// CurriculumJourney tracks the user's 180-day learning journey
type CurriculumJourney struct {
	ID        int64
	UserID    int64
	StartDate time.Time
	IsActive  bool
	CreatedAt time.Time
}

// LessonSession tracks a daily lesson completion
type LessonSession struct {
	ID              int64
	UserID          int64
	SessionDate     time.Time
	DayNumber       int
	PhaseID         int
	CardsReviewed   int
	CardsCorrect    int
	NewCardsLearned int
	XPEarned        int
	CompletedAt     sql.NullTime
	CreatedAt       time.Time
}

// CurriculumPhase defines a phase in the 180-day journey
type CurriculumPhase struct {
	ID              int
	Name            string
	Description     string
	StartDay        int
	EndDay          int
	NewCardsPerDay  int
	TargetIslands   []int64
	ModeWeights     ModeWeights
}

// ModeWeights defines the distribution of practice modes
type ModeWeights struct {
	Standard int // percentage
	Reverse  int // percentage
	Typing   int // percentage
}

// LessonCard represents a card with its assigned practice mode
type LessonCard struct {
	CardWithProgress
	Mode        string // "standard", "reverse", "typing"
	IsNew       bool
	IsSongVocab bool   // true if card is from a song
	SongTitle   string // song title for display (e.g., "Callaíta")
}

// DailyLesson represents the structured lesson for a day
type DailyLesson struct {
	DayNumber      int
	Phase          *CurriculumPhase
	Cards          []LessonCard
	EstimatedMins  int
	DueReviewCount int
	NewCardCount   int
	SuggestedSong  *Song // Song to suggest for full lesson (if any)
}

// JourneyHomeData holds all data for the journey home page
type JourneyHomeData struct {
	DayNumber      int
	TotalDays      int
	CurrentPhase   *CurriculumPhase
	TodayCompleted bool
	TodayStats     string
	DueCount       int
	NewCount       int
	EstimatedMins  int
	Streak         int
	TotalXP        int
}

// CardResult tracks the result of a single card review in a lesson
type CardResult struct {
	CardID      int64
	Term        string
	Translation string
	Rating      Rating
	TimeSpentMs int64
	Mode        string // standard, reverse, typing
	WasCorrect  bool   // rating >= 3
	IsNew       bool
}

// RatingLabel returns a human-readable label for a rating
func (r Rating) Label() string {
	switch r {
	case RatingAgain:
		return "Again"
	case RatingHard:
		return "Hard"
	case RatingGood:
		return "Good"
	case RatingEasy:
		return "Easy"
	default:
		return "Unknown"
	}
}

// LessonSummary holds detailed data for the lesson completion screen
type LessonSummary struct {
	DayNumber      int
	Phase          *CurriculumPhase
	TotalCards     int
	CorrectCount   int
	Accuracy       int
	TotalTimeMs    int64
	AvgTimePerCard int64
	XPEarned       int
	NewLearned     int
	CardResults    []CardResult
	Struggles      []CardResult // Cards with rating 1-2
	Achievements   []Achievement
	Message        string
}

// QuestionType enum for different question formats
type QuestionType string

const (
	QuestionMCQ          QuestionType = "mcq"
	QuestionFillBlank    QuestionType = "fill_blank"
	QuestionSentenceBuild QuestionType = "sentence_build"
)

// Question represents a generated question for a card
type Question struct {
	ID           int64
	CardID       int64
	QuestionType QuestionType
	QuestionData string // JSON with type-specific data
	CreatedAt    time.Time
}

// MCQData represents multiple choice question data
type MCQData struct {
	Stem         string   `json:"stem"`          // The question text
	Options      []string `json:"options"`       // 4 answer options
	CorrectIndex int      `json:"correct_index"` // Index of correct answer (0-3)
	Explanation  string   `json:"explanation"`   // Why the answer is correct
}

// FillBlankData represents fill-in-the-blank question data
type FillBlankData struct {
	Sentence      string `json:"sentence"`       // Sentence with ____ for blank
	BlankPosition int    `json:"blank_position"` // Word index of blank
	Answer        string `json:"answer"`         // Correct answer
	Hint          string `json:"hint"`           // Optional hint
	Context       string `json:"context"`        // English translation
}

// SentenceBuildData represents sentence building question data
type SentenceBuildData struct {
	TargetSentence string   `json:"target_sentence"` // Correct sentence
	WordBank       []string `json:"word_bank"`       // Shuffled words
	Translation    string   `json:"translation"`     // English translation
	Hint           string   `json:"hint"`            // Optional hint
}

// GrammarRule represents a grammar explanation
type GrammarRule struct {
	ID              int64
	RuleKey         string // e.g., "present_tense_ar"
	Title           string // e.g., "Present Tense: -AR Verbs"
	Explanation     string
	Examples        string // JSON array
	RelatedCards    string // JSON array of card IDs
	DifficultyLevel int    // 1=basic, 2=intermediate, 3=advanced
	CreatedAt       time.Time
}

// GrammarExample represents a single grammar example
type GrammarExample struct {
	Spanish string `json:"spanish"`
	English string `json:"english"`
}

// GrammarTip is a short grammar tip for pre-lesson display
type GrammarTip struct {
	RuleKey     string
	Title       string
	ShortExplan string
	Examples    []GrammarExample
}

// =============================================
// SONG LESSONS MODELS
// =============================================

// Song represents a music track for learning
type Song struct {
	ID              int64
	YouTubeID       string
	GeniusID        sql.NullInt64
	Title           string
	Artist          string
	Album           string
	Difficulty      int // 1=beginner, 2=intermediate, 3=advanced
	DurationSeconds int
	ThumbnailURL    string
	AudioPath       string // Local path to MP3 file (served at /songs/audio/)
	CreatedAt       time.Time
	// Joined data
	Lines      []SongLine
	Vocabulary []SongVocab
}

// SongLine represents a single lyric line with timestamps
type SongLine struct {
	ID          int64
	SongID      int64
	LineNumber  int
	StartTimeMs int
	EndTimeMs   int
	SpanishText string
	EnglishText string
}

// SongVocab links song vocabulary to flashcard system
type SongVocab struct {
	ID          int64
	SongID      int64
	CardID      sql.NullInt64
	Word        string
	Translation string
	IsKeyVocab  bool
	// Joined data
	Card *Card // Linked flashcard if exists
}

// SongProgress tracks user's progress on a song (FSRS-based)
type SongProgress struct {
	ID                int64
	UserID            int64
	SongID            int64
	Stability         float64
	Difficulty        float64
	Reps              int
	Lapses            int
	State             CardState // reusing CardState
	Due               sql.NullTime
	LastReview        sql.NullTime
	VocabComplete     bool
	LyricsComplete    bool
	ListeningComplete bool
	TotalListens      int
}

// SongSession tracks a single song lesson session
type SongSession struct {
	ID              int64
	UserID          int64
	SongID          int64
	SessionDate     time.Time
	Mode            SongMode
	VocabReviewed   int
	VocabCorrect    int
	LinesStudied    int
	BlanksCorrect   int
	BlanksTotal     int
	XPEarned        int
	CompletedAt     sql.NullTime
	CreatedAt       time.Time
}

// SongMode enum for song lesson modes
type SongMode string

const (
	SongModeVocab     SongMode = "vocab"
	SongModeLyrics    SongMode = "lyrics"
	SongModeListening SongMode = "listening"
	SongModeFull      SongMode = "full"
)

// SongPhase enum for lesson phases
type SongPhase string

const (
	SongPhaseVocabPreview  SongPhase = "vocab_preview"
	SongPhaseFirstListen   SongPhase = "first_listen"
	SongPhaseLineBreakdown SongPhase = "line_breakdown"
	SongPhaseFillBlanks    SongPhase = "fill_blanks"
	SongPhaseFinalListen   SongPhase = "final_listen"
	SongPhaseComplete      SongPhase = "complete"
)

// SongWithProgress combines song with user progress
type SongWithProgress struct {
	Song
	Progress *SongProgress
}

// SongVocabCard represents a vocab card for song lessons
type SongVocabCard struct {
	SongVocab
	Mode string // standard, reverse
}

// SongBlank represents a fill-in-the-blank for listening mode
type SongBlank struct {
	LineID     int64
	Line       *SongLine
	BlankWord  string
	BlankIndex int // Word position in line
	UserAnswer string
	IsCorrect  bool
}

// SongLesson represents a complete song lesson flow
type SongLesson struct {
	Song          *Song
	Progress      *SongProgress
	CurrentPhase  SongPhase
	CurrentIndex  int // Index within current phase
	VocabCards    []SongVocabCard
	Blanks        []SongBlank
	EstimatedMins int
}

// SongLessonSummary holds completion data
type SongLessonSummary struct {
	Song          *Song
	Mode          SongMode
	VocabReviewed int
	VocabCorrect  int
	LinesStudied  int
	BlanksCorrect int
	BlanksTotal   int
	Accuracy      int
	XPEarned      int
	Message       string
	Achievements  []Achievement
}

// SongHomeData holds data for song lesson home/browse
type SongHomeData struct {
	AvailableSongs    []SongWithProgress
	DueSongs          []SongWithProgress
	TotalSongsLearned int
}

// DifficultyLabel returns a human-readable label for song difficulty
func (s *Song) DifficultyLabel() string {
	switch s.Difficulty {
	case 1:
		return "Beginner"
	case 2:
		return "Intermediate"
	case 3:
		return "Advanced"
	default:
		return "Unknown"
	}
}

// DifficultyClass returns a CSS class for song difficulty
func (s *Song) DifficultyClass() string {
	switch s.Difficulty {
	case 1:
		return "difficulty-beginner"
	case 2:
		return "difficulty-intermediate"
	case 3:
		return "difficulty-advanced"
	default:
		return ""
	}
}

// =============================================
// PROGRESS OVERVIEW MODELS
// =============================================

// IslandProgress represents progress for a single island
type IslandProgress struct {
	IslandID      int64
	IslandName    string
	IslandIcon    string
	TotalCards    int
	LearnedCards  int
	MasteredCards int
}

// ProgressOverviewStats contains aggregate learning statistics
type ProgressOverviewStats struct {
	TotalCards      int
	LearnedCount    int
	LearningCount   int
	ReviewCount     int
	RelearningCount int
	MasteredCount   int
	TotalReviews    int
}

// ProgressOverviewData holds all data for the progress overview page
type ProgressOverviewData struct {
	Stats           *ProgressOverviewStats
	Islands         []IslandProgress
	RecentWords     []CardWithProgress
	RecentLessons   []LessonSession
	TotalXP         int
	CurrentStreak   int
}
