package service

import (
	"database/sql"
	"math/rand"
	"strings"
	"time"

	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// SongService handles song lesson logic
type SongService struct{}

// NewSongService creates a new song service
func NewSongService() *SongService {
	return &SongService{}
}

// GetSongHomeData builds data for the song lessons browse page
func (s *SongService) GetSongHomeData(userID int64) (*models.SongHomeData, error) {
	// Get all songs with user progress
	songs, err := repository.GetSongsWithProgress(userID)
	if err != nil {
		return nil, err
	}

	// Get due songs
	dueSongs, err := repository.GetDueSongs(userID, 10)
	if err != nil {
		return nil, err
	}

	// Count songs learned
	learned, err := repository.CountSongsLearned(userID)
	if err != nil {
		learned = 0
	}

	return &models.SongHomeData{
		AvailableSongs:    songs,
		DueSongs:          dueSongs,
		TotalSongsLearned: learned,
	}, nil
}

// BuildSongLesson creates a song lesson for a specific mode
func (s *SongService) BuildSongLesson(userID, songID int64, mode models.SongMode) (*models.SongLesson, error) {
	// Get song with all details
	song, err := repository.GetSongWithDetails(songID)
	if err != nil {
		return nil, err
	}

	// Ensure song vocab is linked to flashcards for daily lesson integration
	if err := EnsureSongVocabCards(userID, songID); err != nil {
		// Log but don't fail - song lesson can proceed without daily integration
		_ = err
	}

	// Get or create progress
	progress, err := repository.GetOrCreateSongProgress(userID, songID)
	if err != nil {
		return nil, err
	}

	// Build vocab cards (key vocabulary only)
	vocabCards := s.buildVocabCards(song)

	// Build fill-in-the-blanks
	blanks := s.buildFillBlanks(song, 8)

	// Determine starting phase based on mode
	var startPhase models.SongPhase
	switch mode {
	case models.SongModeVocab:
		startPhase = models.SongPhaseVocabPreview
	case models.SongModeLyrics:
		startPhase = models.SongPhaseLineBreakdown
	case models.SongModeListening:
		startPhase = models.SongPhaseFillBlanks
	default: // full
		startPhase = models.SongPhaseVocabPreview
	}

	// Estimate time
	estimatedMins := 3 // base time for video
	estimatedMins += len(vocabCards) / 2
	estimatedMins += len(song.Lines) / 4
	estimatedMins += len(blanks) / 2

	return &models.SongLesson{
		Song:          song,
		Progress:      progress,
		CurrentPhase:  startPhase,
		CurrentIndex:  0,
		VocabCards:    vocabCards,
		Blanks:        blanks,
		EstimatedMins: estimatedMins,
	}, nil
}

// buildVocabCards creates vocab flashcards from song vocabulary
func (s *SongService) buildVocabCards(song *models.Song) []models.SongVocabCard {
	var cards []models.SongVocabCard

	for _, v := range song.Vocabulary {
		if !v.IsKeyVocab {
			continue
		}
		mode := "standard"
		if rand.Intn(100) < 30 {
			mode = "reverse"
		}
		cards = append(cards, models.SongVocabCard{
			SongVocab: v,
			Mode:      mode,
		})
	}

	return cards
}

// buildFillBlanks generates fill-in-the-blank questions from song lines
func (s *SongService) buildFillBlanks(song *models.Song, count int) []models.SongBlank {
	if len(song.Lines) == 0 {
		return nil
	}

	// Build a set of vocabulary words for blanking
	vocabWords := make(map[string]bool)
	for _, v := range song.Vocabulary {
		vocabWords[strings.ToLower(v.Word)] = true
	}

	var candidates []models.SongBlank

	for i := range song.Lines {
		line := &song.Lines[i]
		words := strings.Fields(line.SpanishText)
		if len(words) < 3 {
			continue
		}

		// Find words that are in vocabulary
		for idx, word := range words {
			cleanWord := strings.ToLower(strings.Trim(word, ".,!?¿¡"))
			if vocabWords[cleanWord] {
				candidates = append(candidates, models.SongBlank{
					LineID:     line.ID,
					Line:       line,
					BlankWord:  cleanWord,
					BlankIndex: idx,
				})
			}
		}
	}

	// If not enough vocab words, pick random words
	if len(candidates) < count {
		for i := range song.Lines {
			if len(candidates) >= count*2 {
				break
			}
			line := &song.Lines[i]
			words := strings.Fields(line.SpanishText)
			if len(words) < 3 {
				continue
			}
			// Pick a random word (not first or last)
			idx := rand.Intn(len(words)-2) + 1
			word := strings.ToLower(strings.Trim(words[idx], ".,!?¿¡"))
			if len(word) >= 3 { // Skip very short words
				candidates = append(candidates, models.SongBlank{
					LineID:     line.ID,
					Line:       line,
					BlankWord:  word,
					BlankIndex: idx,
				})
			}
		}
	}

	// Shuffle and limit
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if len(candidates) > count {
		candidates = candidates[:count]
	}

	return candidates
}

// CheckBlankAnswer checks if the user's answer is correct
func (s *SongService) CheckBlankAnswer(blank *models.SongBlank, answer string) bool {
	answer = strings.ToLower(strings.TrimSpace(answer))
	correct := strings.ToLower(blank.BlankWord)

	// Exact match
	if answer == correct {
		return true
	}

	// Lenient match - remove accents for comparison
	answerClean := removeAccents(answer)
	correctClean := removeAccents(correct)

	return answerClean == correctClean
}

// removeAccents removes Spanish accents for lenient matching
func removeAccents(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"ü", "u", "ñ", "n",
	)
	return replacer.Replace(s)
}

// CalculateSongXP calculates XP for song lesson completion
func CalculateSongXP(mode models.SongMode, vocabCorrect, vocabTotal, blanksCorrect, blanksTotal int) int {
	xp := 0

	// Base XP for completing a song lesson
	switch mode {
	case models.SongModeVocab:
		xp = 10
	case models.SongModeLyrics:
		xp = 15
	case models.SongModeListening:
		xp = 20
	case models.SongModeFull:
		xp = 30
	}

	// Vocab bonus
	if vocabTotal > 0 {
		accuracy := float64(vocabCorrect) / float64(vocabTotal)
		xp += int(accuracy * 10)
	}

	// Blanks bonus
	if blanksTotal > 0 {
		accuracy := float64(blanksCorrect) / float64(blanksTotal)
		xp += int(accuracy * 15)
	}

	return xp
}

// UpdateSongProgressAfterLesson updates song progress after completing a lesson
func (s *SongService) UpdateSongProgressAfterLesson(
	userID, songID int64,
	mode models.SongMode,
	vocabCorrect, vocabTotal, blanksCorrect, blanksTotal int,
) error {
	progress, err := repository.GetOrCreateSongProgress(userID, songID)
	if err != nil {
		return err
	}

	// Calculate overall performance
	totalItems := vocabTotal + blanksTotal
	totalCorrect := vocabCorrect + blanksCorrect
	accuracy := 0.0
	if totalItems > 0 {
		accuracy = float64(totalCorrect) / float64(totalItems)
	}

	// Update completion flags based on mode
	switch mode {
	case models.SongModeVocab:
		progress.VocabComplete = true
	case models.SongModeLyrics:
		progress.LyricsComplete = true
	case models.SongModeListening:
		progress.ListeningComplete = true
	case models.SongModeFull:
		progress.VocabComplete = true
		progress.LyricsComplete = true
		progress.ListeningComplete = true
	}

	// Update FSRS-like progress
	progress.Reps++
	progress.LastReview = toNullTime(time.Now())

	// Simple stability update based on accuracy
	if accuracy >= 0.8 {
		progress.Stability = progress.Stability*1.5 + 1
		progress.State = models.StateReview
	} else if accuracy >= 0.6 {
		progress.Stability = progress.Stability*1.2 + 0.5
		progress.State = models.StateLearning
	} else {
		progress.Stability = progress.Stability * 0.8
		progress.Lapses++
		progress.State = models.StateRelearning
	}

	// Schedule next review
	daysUntilReview := int(progress.Stability)
	if daysUntilReview < 1 {
		daysUntilReview = 1
	}
	if daysUntilReview > 30 {
		daysUntilReview = 30
	}
	nextDue := time.Now().AddDate(0, 0, daysUntilReview)
	progress.Due = toNullTime(nextDue)

	return repository.UpsertSongProgress(progress)
}

// GetSongMotivationalMessage returns a message based on performance
func GetSongMotivationalMessage(accuracy int) string {
	if accuracy >= 90 {
		return "Amazing! You really know this song!"
	}
	if accuracy >= 80 {
		return "Great listening skills!"
	}
	if accuracy >= 70 {
		return "Good job! Keep practicing!"
	}
	if accuracy >= 60 {
		return "Nice effort! Listen again to catch more."
	}
	return "Keep listening! You'll get better with practice."
}

// RenderBlankLine creates a display string with the blank word replaced
func RenderBlankLine(spanishText string, blankIndex int) string {
	words := strings.Fields(spanishText)
	if blankIndex < 0 || blankIndex >= len(words) {
		return spanishText
	}
	words[blankIndex] = "____"
	return strings.Join(words, " ")
}

// toNullTime converts a time to sql.NullTime
func toNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

// GetNextPhase returns the next phase in the song lesson flow
func GetNextPhase(current models.SongPhase, mode models.SongMode) models.SongPhase {
	switch current {
	case models.SongPhaseVocabPreview:
		if mode == models.SongModeVocab {
			return models.SongPhaseComplete
		}
		return models.SongPhaseFirstListen

	case models.SongPhaseFirstListen:
		return models.SongPhaseLineBreakdown

	case models.SongPhaseLineBreakdown:
		if mode == models.SongModeLyrics {
			return models.SongPhaseComplete
		}
		return models.SongPhaseFillBlanks

	case models.SongPhaseFillBlanks:
		if mode == models.SongModeListening {
			return models.SongPhaseComplete
		}
		return models.SongPhaseFinalListen

	case models.SongPhaseFinalListen:
		return models.SongPhaseComplete

	default:
		return models.SongPhaseComplete
	}
}

// ShouldSkipPhase determines if a phase should be skipped based on mode
func ShouldSkipPhase(phase models.SongPhase, mode models.SongMode) bool {
	switch mode {
	case models.SongModeVocab:
		// Only vocab preview
		return phase != models.SongPhaseVocabPreview && phase != models.SongPhaseComplete
	case models.SongModeLyrics:
		// Skip vocab and listening
		return phase == models.SongPhaseVocabPreview ||
			phase == models.SongPhaseFirstListen ||
			phase == models.SongPhaseFillBlanks ||
			phase == models.SongPhaseFinalListen
	case models.SongModeListening:
		// Only fill blanks
		return phase != models.SongPhaseFillBlanks && phase != models.SongPhaseComplete
	default:
		// Full mode - no skipping
		return false
	}
}

// EnsureSongVocabCards creates flashcards for song vocabulary that don't have cards yet
// This allows song vocab to be mixed into the daily lesson flow
func EnsureSongVocabCards(userID, songID int64) error {
	// Get song for title
	song, err := repository.GetSong(songID)
	if err != nil {
		return err
	}

	// Get vocabulary without linked cards
	unlinked, err := repository.GetUnlinkedSongVocab(songID)
	if err != nil {
		return err
	}

	// Create a card for each unlinked vocab
	for _, v := range unlinked {
		card := &models.Card{
			Term:         v.Word,
			Translation:  v.Translation,
			Source:       "song",
			SourceSongID: sql.NullInt64{Int64: songID, Valid: true},
			Notes:        "From song: " + song.Title,
		}

		if err := repository.CreateCard(card); err != nil {
			return err
		}

		// Link the vocab entry to the new card
		if err := repository.LinkSongVocabToCard(v.ID, card.ID); err != nil {
			return err
		}
	}

	return nil
}
