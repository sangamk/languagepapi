package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/dhowden/tag"

	"languagepapi/components"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
	"languagepapi/internal/service"
)

var (
	// In-memory song lesson storage
	songLessons     = make(map[int64]*models.SongLesson)
	songStats       = make(map[int64]*songSessionStats)
	songSessions    = make(map[int64]*models.SongSession)
	songLessonsLock sync.RWMutex
)

type songSessionStats struct {
	VocabReviewed int
	VocabCorrect  int
	LinesStudied  int
	BlanksCorrect int
	BlanksTotal   int
	XPEarned      int
	StartTime     time.Time
}

var songService = service.NewSongService()

// SongsPath is the path to the songs directory (set from main.go)
var SongsPath = "./songs"

// HandleAlbumArt extracts and serves album art from audio file metadata
func HandleAlbumArt(w http.ResponseWriter, r *http.Request) {
	// Get the audio filename from the path (e.g., /audio/cover/01.%20Moscow%20Mule.mp3)
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "missing filename", http.StatusBadRequest)
		return
	}

	// Open the audio file
	audioPath := filepath.Join(SongsPath, filename)
	f, err := os.Open(audioPath)
	if err != nil {
		http.Error(w, "audio file not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Read metadata
	m, err := tag.ReadFrom(f)
	if err != nil {
		http.Error(w, "could not read metadata", http.StatusInternalServerError)
		return
	}

	// Get embedded picture
	pic := m.Picture()
	if pic == nil {
		http.Error(w, "no album art found", http.StatusNotFound)
		return
	}

	// Set content type and serve
	w.Header().Set("Content-Type", pic.MIMEType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	w.Write(pic.Data)
}

// HandleSongHome renders the song lessons browse page
func HandleSongHome(w http.ResponseWriter, r *http.Request) {
	data, err := songService.GetSongHomeData(defaultUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	components.SongHome(data).Render(r.Context(), w)
}

// HandleSongDetail renders the song detail page with mode selection
func HandleSongDetail(w http.ResponseWriter, r *http.Request) {
	songID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid song id", http.StatusBadRequest)
		return
	}

	song, err := repository.GetSongWithDetails(songID)
	if err != nil {
		http.Error(w, "song not found", http.StatusNotFound)
		return
	}

	progress, _ := repository.GetSongProgress(defaultUserID, songID)

	components.SongDetail(song, progress).Render(r.Context(), w)
}

// HandleSongStart starts a song lesson
func HandleSongStart(w http.ResponseWriter, r *http.Request) {
	songID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid song id", http.StatusBadRequest)
		return
	}

	mode := models.SongMode(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = models.SongModeFull
	}

	songLessonsLock.Lock()
	defer songLessonsLock.Unlock()

	// Build lesson
	lesson, err := songService.BuildSongLesson(defaultUserID, songID, mode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store in memory
	songLessons[defaultUserID] = lesson
	songStats[defaultUserID] = &songSessionStats{
		StartTime:   time.Now(),
		BlanksTotal: len(lesson.Blanks),
	}

	// Create session in DB
	session := &models.SongSession{
		UserID: defaultUserID,
		SongID: songID,
		Mode:   mode,
	}
	repository.CreateSongSession(session)
	songSessions[defaultUserID] = session

	// Render appropriate first phase
	renderCurrentPhase(w, r, lesson)
}

// HandleSongVocabReview processes a vocab card review in song lesson
func HandleSongVocabReview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil || rating < 1 || rating > 4 {
		http.Error(w, "invalid rating", http.StatusBadRequest)
		return
	}

	songLessonsLock.Lock()
	lesson, exists := songLessons[defaultUserID]
	if !exists {
		songLessonsLock.Unlock()
		http.Redirect(w, r, "/songs", http.StatusSeeOther)
		return
	}

	stats := songStats[defaultUserID]
	stats.VocabReviewed++
	if rating >= 3 {
		stats.VocabCorrect++
		stats.XPEarned += 2
	}

	// Move to next vocab card
	lesson.CurrentIndex++

	// Check if vocab phase is complete
	if lesson.CurrentIndex >= len(lesson.VocabCards) {
		lesson.CurrentIndex = 0
		lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
	}
	songLessonsLock.Unlock()

	renderCurrentPhase(w, r, lesson)
}

// HandleSongNextPhase advances to next phase of song lesson
func HandleSongNextPhase(w http.ResponseWriter, r *http.Request) {
	songLessonsLock.Lock()
	lesson, exists := songLessons[defaultUserID]
	if !exists {
		songLessonsLock.Unlock()
		http.Redirect(w, r, "/songs", http.StatusSeeOther)
		return
	}

	// Increment listen count when moving past first listen
	if lesson.CurrentPhase == models.SongPhaseFirstListen {
		repository.IncrementSongListenCount(defaultUserID, lesson.Song.ID)
	}

	lesson.CurrentIndex = 0
	lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
	songLessonsLock.Unlock()

	renderCurrentPhase(w, r, lesson)
}

// HandleSongNextLine advances to next line in breakdown phase
func HandleSongNextLine(w http.ResponseWriter, r *http.Request) {
	songLessonsLock.Lock()
	lesson, exists := songLessons[defaultUserID]
	if !exists {
		songLessonsLock.Unlock()
		http.Redirect(w, r, "/songs", http.StatusSeeOther)
		return
	}

	stats := songStats[defaultUserID]
	stats.LinesStudied++

	lesson.CurrentIndex++

	// Check if breakdown phase is complete
	if lesson.CurrentIndex >= len(lesson.Song.Lines) {
		lesson.CurrentIndex = 0
		lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
	}
	songLessonsLock.Unlock()

	renderCurrentPhase(w, r, lesson)
}

// HandleSongSkipLine skips the current line
func HandleSongSkipLine(w http.ResponseWriter, r *http.Request) {
	songLessonsLock.Lock()
	lesson, exists := songLessons[defaultUserID]
	if !exists {
		songLessonsLock.Unlock()
		http.Redirect(w, r, "/songs", http.StatusSeeOther)
		return
	}

	lesson.CurrentIndex++

	// Check if breakdown phase is complete
	if lesson.CurrentIndex >= len(lesson.Song.Lines) {
		lesson.CurrentIndex = 0
		lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
	}
	songLessonsLock.Unlock()

	renderCurrentPhase(w, r, lesson)
}

// HandleSongBlankSubmit checks fill-in-the-blank answer
func HandleSongBlankSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	answer := r.FormValue("answer")

	songLessonsLock.Lock()
	lesson, exists := songLessons[defaultUserID]
	if !exists {
		songLessonsLock.Unlock()
		http.Redirect(w, r, "/songs", http.StatusSeeOther)
		return
	}

	stats := songStats[defaultUserID]
	currentIdx := lesson.CurrentIndex

	if currentIdx < len(lesson.Blanks) {
		blank := &lesson.Blanks[currentIdx]
		blank.UserAnswer = answer
		blank.IsCorrect = songService.CheckBlankAnswer(blank, answer)

		if blank.IsCorrect {
			stats.BlanksCorrect++
			stats.XPEarned += 3
		}
	}

	lesson.CurrentIndex++

	// Check if blanks phase is complete
	if lesson.CurrentIndex >= len(lesson.Blanks) {
		lesson.CurrentIndex = 0
		lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
	}
	songLessonsLock.Unlock()

	renderCurrentPhase(w, r, lesson)
}

// HandleSongComplete completes the song lesson
func HandleSongComplete(w http.ResponseWriter, r *http.Request) {
	songLessonsLock.Lock()
	lesson, exists := songLessons[defaultUserID]
	if !exists {
		songLessonsLock.Unlock()
		http.Redirect(w, r, "/songs", http.StatusSeeOther)
		return
	}

	stats := songStats[defaultUserID]
	session := songSessions[defaultUserID]

	// Calculate final XP
	mode := getSongMode(lesson)
	stats.XPEarned += service.CalculateSongXP(mode, stats.VocabCorrect, stats.VocabReviewed, stats.BlanksCorrect, stats.BlanksTotal)

	// Update session in DB
	if session != nil {
		repository.UpdateSongSession(session.ID, stats.VocabReviewed, stats.VocabCorrect, stats.LinesStudied, stats.BlanksCorrect, stats.BlanksTotal, stats.XPEarned)
		repository.CompleteSongSession(session.ID)
	}

	// Update song progress
	songService.UpdateSongProgressAfterLesson(defaultUserID, lesson.Song.ID, mode, stats.VocabCorrect, stats.VocabReviewed, stats.BlanksCorrect, stats.BlanksTotal)

	// Update user XP
	repository.UpdateUserXP(defaultUserID, stats.XPEarned)

	// Calculate accuracy
	accuracy := 0
	totalItems := stats.VocabReviewed + stats.BlanksTotal
	if totalItems > 0 {
		accuracy = (stats.VocabCorrect + stats.BlanksCorrect) * 100 / totalItems
	}

	// Build summary
	summary := &models.SongLessonSummary{
		Song:          lesson.Song,
		Mode:          mode,
		VocabReviewed: stats.VocabReviewed,
		VocabCorrect:  stats.VocabCorrect,
		LinesStudied:  stats.LinesStudied,
		BlanksCorrect: stats.BlanksCorrect,
		BlanksTotal:   stats.BlanksTotal,
		Accuracy:      accuracy,
		XPEarned:      stats.XPEarned,
		Message:       service.GetSongMotivationalMessage(accuracy),
	}

	// Clear session data
	delete(songLessons, defaultUserID)
	delete(songStats, defaultUserID)
	delete(songSessions, defaultUserID)
	songLessonsLock.Unlock()

	components.SongComplete(summary).Render(r.Context(), w)
}

// getSongMode determines the mode from the lesson
func getSongMode(lesson *models.SongLesson) models.SongMode {
	// Infer from what's available
	hasVocab := len(lesson.VocabCards) > 0
	hasBlanks := len(lesson.Blanks) > 0
	hasLines := len(lesson.Song.Lines) > 0

	if hasVocab && hasLines && hasBlanks {
		return models.SongModeFull
	}
	if hasVocab && !hasLines && !hasBlanks {
		return models.SongModeVocab
	}
	if hasLines && !hasVocab && !hasBlanks {
		return models.SongModeLyrics
	}
	if hasBlanks && !hasVocab && !hasLines {
		return models.SongModeListening
	}
	return models.SongModeFull
}

// renderCurrentPhase renders the appropriate template for the current phase
func renderCurrentPhase(w http.ResponseWriter, r *http.Request, lesson *models.SongLesson) {
	switch lesson.CurrentPhase {
	case models.SongPhaseVocabPreview:
		if len(lesson.VocabCards) == 0 || lesson.CurrentIndex >= len(lesson.VocabCards) {
			lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
			renderCurrentPhase(w, r, lesson)
			return
		}
		card := &lesson.VocabCards[lesson.CurrentIndex]
		components.SongVocabPhase(lesson, lesson.CurrentIndex+1, len(lesson.VocabCards), card).Render(r.Context(), w)

	case models.SongPhaseFirstListen:
		components.SongFirstListen(lesson).Render(r.Context(), w)

	case models.SongPhaseLineBreakdown:
		if len(lesson.Song.Lines) == 0 || lesson.CurrentIndex >= len(lesson.Song.Lines) {
			lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
			renderCurrentPhase(w, r, lesson)
			return
		}
		components.SongLineBreakdown(lesson, lesson.CurrentIndex).Render(r.Context(), w)

	case models.SongPhaseFillBlanks:
		if len(lesson.Blanks) == 0 || lesson.CurrentIndex >= len(lesson.Blanks) {
			lesson.CurrentPhase = service.GetNextPhase(lesson.CurrentPhase, getSongMode(lesson))
			renderCurrentPhase(w, r, lesson)
			return
		}
		blank := &lesson.Blanks[lesson.CurrentIndex]
		components.SongFillBlanks(lesson, blank, lesson.CurrentIndex+1, len(lesson.Blanks)).Render(r.Context(), w)

	case models.SongPhaseFinalListen:
		components.SongFinalListen(lesson).Render(r.Context(), w)

	case models.SongPhaseComplete:
		// Trigger completion handler
		HandleSongComplete(w, r)

	default:
		http.Redirect(w, r, "/songs", http.StatusSeeOther)
	}
}

// HandleFetchLyrics fetches lyrics from lrclib.net for a song
func HandleFetchLyrics(w http.ResponseWriter, r *http.Request) {
	songID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid song id", http.StatusBadRequest)
		return
	}

	lyricsService := service.NewLyricsService()
	err = lyricsService.FetchAndStoreLyrics(songID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}
