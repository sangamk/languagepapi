package repository

import (
	"database/sql"
	"time"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetSong retrieves a song by ID
func GetSong(id int64) (*models.Song, error) {
	s := &models.Song{}
	err := db.DB.QueryRow(`
		SELECT id, COALESCE(youtube_id, ''), genius_id, title, artist, COALESCE(album, ''),
		       difficulty, COALESCE(duration_seconds, 0), COALESCE(thumbnail_url, ''),
		       COALESCE(audio_path, ''), created_at
		FROM songs
		WHERE id = ?
	`, id).Scan(&s.ID, &s.YouTubeID, &s.GeniusID, &s.Title, &s.Artist, &s.Album,
		&s.Difficulty, &s.DurationSeconds, &s.ThumbnailURL, &s.AudioPath, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetSongByYouTubeID retrieves a song by YouTube ID
func GetSongByYouTubeID(youtubeID string) (*models.Song, error) {
	s := &models.Song{}
	err := db.DB.QueryRow(`
		SELECT id, COALESCE(youtube_id, ''), genius_id, title, artist, COALESCE(album, ''),
		       difficulty, COALESCE(duration_seconds, 0), COALESCE(thumbnail_url, ''),
		       COALESCE(audio_path, ''), created_at
		FROM songs
		WHERE youtube_id = ?
	`, youtubeID).Scan(&s.ID, &s.YouTubeID, &s.GeniusID, &s.Title, &s.Artist, &s.Album,
		&s.Difficulty, &s.DurationSeconds, &s.ThumbnailURL, &s.AudioPath, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetSongWithDetails retrieves song with lines and vocabulary
func GetSongWithDetails(id int64) (*models.Song, error) {
	song, err := GetSong(id)
	if err != nil {
		return nil, err
	}

	// Get lines
	lines, err := GetSongLines(id)
	if err != nil {
		return nil, err
	}
	song.Lines = lines

	// Get vocabulary
	vocab, err := GetSongVocabulary(id, false)
	if err != nil {
		return nil, err
	}
	song.Vocabulary = vocab

	return song, nil
}

// GetAllSongs retrieves all songs with optional difficulty filter
func GetAllSongs(difficulty int) ([]models.Song, error) {
	var query string
	var args []interface{}

	if difficulty > 0 {
		query = `
			SELECT id, COALESCE(youtube_id, ''), genius_id, title, artist, COALESCE(album, ''),
			       difficulty, COALESCE(duration_seconds, 0), COALESCE(thumbnail_url, ''),
			       COALESCE(audio_path, ''), created_at
			FROM songs
			WHERE difficulty = ?
			ORDER BY title ASC
		`
		args = append(args, difficulty)
	} else {
		query = `
			SELECT id, COALESCE(youtube_id, ''), genius_id, title, artist, COALESCE(album, ''),
			       difficulty, COALESCE(duration_seconds, 0), COALESCE(thumbnail_url, ''),
			       COALESCE(audio_path, ''), created_at
			FROM songs
			ORDER BY difficulty ASC, title ASC
		`
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var s models.Song
		if err := rows.Scan(&s.ID, &s.YouTubeID, &s.GeniusID, &s.Title, &s.Artist, &s.Album,
			&s.Difficulty, &s.DurationSeconds, &s.ThumbnailURL, &s.AudioPath, &s.CreatedAt); err != nil {
			return nil, err
		}
		songs = append(songs, s)
	}
	return songs, rows.Err()
}

// GetSongLines retrieves all lines for a song
func GetSongLines(songID int64) ([]models.SongLine, error) {
	rows, err := db.DB.Query(`
		SELECT id, song_id, line_number, start_time_ms, end_time_ms,
		       spanish_text, english_text
		FROM song_lines
		WHERE song_id = ?
		ORDER BY line_number ASC
	`, songID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []models.SongLine
	for rows.Next() {
		var l models.SongLine
		if err := rows.Scan(&l.ID, &l.SongID, &l.LineNumber, &l.StartTimeMs,
			&l.EndTimeMs, &l.SpanishText, &l.EnglishText); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

// GetSongVocabulary retrieves vocabulary for a song
func GetSongVocabulary(songID int64, keyOnly bool) ([]models.SongVocab, error) {
	query := `
		SELECT id, song_id, card_id, word, translation, is_key_vocab
		FROM song_vocabulary
		WHERE song_id = ?
	`
	if keyOnly {
		query += ` AND is_key_vocab = 1`
	}
	query += ` ORDER BY id ASC`

	rows, err := db.DB.Query(query, songID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vocab []models.SongVocab
	for rows.Next() {
		var v models.SongVocab
		var isKey int
		if err := rows.Scan(&v.ID, &v.SongID, &v.CardID, &v.Word, &v.Translation, &isKey); err != nil {
			return nil, err
		}
		v.IsKeyVocab = isKey == 1
		vocab = append(vocab, v)
	}
	return vocab, rows.Err()
}

// GetSongProgress retrieves user's progress on a song
func GetSongProgress(userID, songID int64) (*models.SongProgress, error) {
	p := &models.SongProgress{}
	var vocabComplete, lyricsComplete, listeningComplete int
	err := db.DB.QueryRow(`
		SELECT id, user_id, song_id, stability, difficulty, reps, lapses,
		       state, due, last_review, vocab_complete, lyrics_complete,
		       listening_complete, total_listens
		FROM song_progress
		WHERE user_id = ? AND song_id = ?
	`, userID, songID).Scan(
		&p.ID, &p.UserID, &p.SongID, &p.Stability, &p.Difficulty,
		&p.Reps, &p.Lapses, &p.State, &p.Due, &p.LastReview,
		&vocabComplete, &lyricsComplete, &listeningComplete, &p.TotalListens,
	)
	if err != nil {
		return nil, err
	}
	p.VocabComplete = vocabComplete == 1
	p.LyricsComplete = lyricsComplete == 1
	p.ListeningComplete = listeningComplete == 1
	return p, nil
}

// UpsertSongProgress creates or updates song progress
func UpsertSongProgress(p *models.SongProgress) error {
	vocabComplete := 0
	if p.VocabComplete {
		vocabComplete = 1
	}
	lyricsComplete := 0
	if p.LyricsComplete {
		lyricsComplete = 1
	}
	listeningComplete := 0
	if p.ListeningComplete {
		listeningComplete = 1
	}

	_, err := db.DB.Exec(`
		INSERT INTO song_progress (user_id, song_id, stability, difficulty, reps, lapses,
		                           state, due, last_review, vocab_complete, lyrics_complete,
		                           listening_complete, total_listens)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, song_id) DO UPDATE SET
			stability = excluded.stability,
			difficulty = excluded.difficulty,
			reps = excluded.reps,
			lapses = excluded.lapses,
			state = excluded.state,
			due = excluded.due,
			last_review = excluded.last_review,
			vocab_complete = excluded.vocab_complete,
			lyrics_complete = excluded.lyrics_complete,
			listening_complete = excluded.listening_complete,
			total_listens = excluded.total_listens
	`, p.UserID, p.SongID, p.Stability, p.Difficulty, p.Reps, p.Lapses,
		p.State, p.Due, p.LastReview, vocabComplete, lyricsComplete,
		listeningComplete, p.TotalListens)
	return err
}

// GetDueSongs returns songs due for review
func GetDueSongs(userID int64, limit int) ([]models.SongWithProgress, error) {
	rows, err := db.DB.Query(`
		SELECT s.id, COALESCE(s.youtube_id, ''), s.genius_id, s.title, s.artist, COALESCE(s.album, ''),
		       s.difficulty, COALESCE(s.duration_seconds, 0), COALESCE(s.thumbnail_url, ''),
		       COALESCE(s.audio_path, ''), s.created_at,
		       p.id, p.stability, p.difficulty, p.reps, p.lapses, p.state,
		       p.due, p.last_review, p.vocab_complete, p.lyrics_complete,
		       p.listening_complete, p.total_listens
		FROM songs s
		JOIN song_progress p ON s.id = p.song_id
		WHERE p.user_id = ? AND p.due <= CURRENT_TIMESTAMP AND p.state != 'new'
		ORDER BY p.due ASC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []models.SongWithProgress
	for rows.Next() {
		var swp models.SongWithProgress
		var prog models.SongProgress
		var vocabComplete, lyricsComplete, listeningComplete int
		if err := rows.Scan(
			&swp.ID, &swp.YouTubeID, &swp.GeniusID, &swp.Title, &swp.Artist, &swp.Album,
			&swp.Difficulty, &swp.DurationSeconds, &swp.ThumbnailURL, &swp.AudioPath, &swp.CreatedAt,
			&prog.ID, &prog.Stability, &prog.Difficulty, &prog.Reps, &prog.Lapses,
			&prog.State, &prog.Due, &prog.LastReview, &vocabComplete, &lyricsComplete,
			&listeningComplete, &prog.TotalListens,
		); err != nil {
			return nil, err
		}
		prog.VocabComplete = vocabComplete == 1
		prog.LyricsComplete = lyricsComplete == 1
		prog.ListeningComplete = listeningComplete == 1
		prog.UserID = userID
		prog.SongID = swp.ID
		swp.Progress = &prog
		songs = append(songs, swp)
	}
	return songs, rows.Err()
}

// GetSongsWithProgress returns all songs with user progress
func GetSongsWithProgress(userID int64) ([]models.SongWithProgress, error) {
	rows, err := db.DB.Query(`
		SELECT s.id, COALESCE(s.youtube_id, ''), s.genius_id, s.title, s.artist, COALESCE(s.album, ''),
		       s.difficulty, COALESCE(s.duration_seconds, 0), COALESCE(s.thumbnail_url, ''),
		       COALESCE(s.audio_path, ''), s.created_at,
		       p.id, p.stability, p.difficulty, p.reps, p.lapses, p.state,
		       p.due, p.last_review, p.vocab_complete, p.lyrics_complete,
		       p.listening_complete, p.total_listens
		FROM songs s
		LEFT JOIN song_progress p ON s.id = p.song_id AND p.user_id = ?
		ORDER BY s.difficulty ASC, s.title ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []models.SongWithProgress
	for rows.Next() {
		var swp models.SongWithProgress
		var progID sql.NullInt64
		var stability, difficulty sql.NullFloat64
		var reps, lapses, totalListens sql.NullInt64
		var state sql.NullString
		var due, lastReview sql.NullTime
		var vocabComplete, lyricsComplete, listeningComplete sql.NullInt64

		if err := rows.Scan(
			&swp.ID, &swp.YouTubeID, &swp.GeniusID, &swp.Title, &swp.Artist, &swp.Album,
			&swp.Difficulty, &swp.DurationSeconds, &swp.ThumbnailURL, &swp.AudioPath, &swp.CreatedAt,
			&progID, &stability, &difficulty, &reps, &lapses,
			&state, &due, &lastReview, &vocabComplete, &lyricsComplete,
			&listeningComplete, &totalListens,
		); err != nil {
			return nil, err
		}

		if progID.Valid {
			swp.Progress = &models.SongProgress{
				ID:                progID.Int64,
				UserID:            userID,
				SongID:            swp.ID,
				Stability:         stability.Float64,
				Difficulty:        difficulty.Float64,
				Reps:              int(reps.Int64),
				Lapses:            int(lapses.Int64),
				State:             models.CardState(state.String),
				Due:               due,
				LastReview:        lastReview,
				VocabComplete:     vocabComplete.Int64 == 1,
				LyricsComplete:    lyricsComplete.Int64 == 1,
				ListeningComplete: listeningComplete.Int64 == 1,
				TotalListens:      int(totalListens.Int64),
			}
		}
		songs = append(songs, swp)
	}
	return songs, rows.Err()
}

// CreateSongSession creates a new song lesson session
func CreateSongSession(s *models.SongSession) error {
	today := time.Now().Format("2006-01-02")
	result, err := db.DB.Exec(`
		INSERT INTO song_sessions (user_id, song_id, session_date, mode)
		VALUES (?, ?, ?, ?)
	`, s.UserID, s.SongID, today, s.Mode)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	s.ID = id
	return nil
}

// UpdateSongSession updates session stats
func UpdateSongSession(sessionID int64, vocabReviewed, vocabCorrect, linesStudied, blanksCorrect, blanksTotal, xp int) error {
	_, err := db.DB.Exec(`
		UPDATE song_sessions
		SET vocab_reviewed = vocab_reviewed + ?,
		    vocab_correct = vocab_correct + ?,
		    lines_studied = lines_studied + ?,
		    blanks_correct = blanks_correct + ?,
		    blanks_total = blanks_total + ?,
		    xp_earned = xp_earned + ?
		WHERE id = ?
	`, vocabReviewed, vocabCorrect, linesStudied, blanksCorrect, blanksTotal, xp, sessionID)
	return err
}

// CompleteSongSession marks session as complete
func CompleteSongSession(sessionID int64) error {
	_, err := db.DB.Exec(`
		UPDATE song_sessions
		SET completed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, sessionID)
	return err
}

// IncrementSongListenCount increments the listen count
func IncrementSongListenCount(userID, songID int64) error {
	_, err := db.DB.Exec(`
		UPDATE song_progress
		SET total_listens = total_listens + 1
		WHERE user_id = ? AND song_id = ?
	`, userID, songID)
	return err
}

// CountSongsLearned returns count of songs user has studied
func CountSongsLearned(userID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM song_progress
		WHERE user_id = ? AND reps > 0
	`, userID).Scan(&count)
	return count, err
}

// CreateSong inserts a new song
func CreateSong(song *models.Song) error {
	result, err := db.DB.Exec(`
		INSERT INTO songs (youtube_id, genius_id, title, artist, album, difficulty, duration_seconds, thumbnail_url, audio_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, song.YouTubeID, song.GeniusID, song.Title, song.Artist, song.Album, song.Difficulty, song.DurationSeconds, song.ThumbnailURL, song.AudioPath)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	song.ID = id
	return nil
}

// CreateSongLine inserts a song line
func CreateSongLine(line *models.SongLine) error {
	result, err := db.DB.Exec(`
		INSERT INTO song_lines (song_id, line_number, start_time_ms, end_time_ms, spanish_text, english_text)
		VALUES (?, ?, ?, ?, ?, ?)
	`, line.SongID, line.LineNumber, line.StartTimeMs, line.EndTimeMs, line.SpanishText, line.EnglishText)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	line.ID = id
	return nil
}

// CreateSongVocab inserts song vocabulary
func CreateSongVocab(vocab *models.SongVocab) error {
	isKey := 0
	if vocab.IsKeyVocab {
		isKey = 1
	}
	result, err := db.DB.Exec(`
		INSERT INTO song_vocabulary (song_id, card_id, word, translation, is_key_vocab)
		VALUES (?, ?, ?, ?, ?)
	`, vocab.SongID, vocab.CardID, vocab.Word, vocab.Translation, isKey)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	vocab.ID = id
	return nil
}

// GetOrCreateSongProgress gets existing progress or creates initial state
func GetOrCreateSongProgress(userID, songID int64) (*models.SongProgress, error) {
	progress, err := GetSongProgress(userID, songID)
	if err == sql.ErrNoRows {
		// Create initial progress
		progress = &models.SongProgress{
			UserID:     userID,
			SongID:     songID,
			State:      models.StateNew,
			Stability:  0,
			Difficulty: 0,
		}
		err = UpsertSongProgress(progress)
		if err != nil {
			return nil, err
		}
		return progress, nil
	}
	return progress, err
}

// GetUserSongsInProgress returns song IDs that the user has started learning
func GetUserSongsInProgress(userID int64) ([]int64, error) {
	rows, err := db.DB.Query(`
		SELECT song_id FROM song_progress
		WHERE user_id = ? AND reps > 0
		ORDER BY last_review DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		songIDs = append(songIDs, id)
	}
	return songIDs, rows.Err()
}

// LinkSongVocabToCard links a song vocabulary entry to a card
func LinkSongVocabToCard(vocabID, cardID int64) error {
	_, err := db.DB.Exec(`
		UPDATE song_vocabulary SET card_id = ? WHERE id = ?
	`, cardID, vocabID)
	return err
}

// UpdateSongGeniusInfo updates a song with Genius API data
func UpdateSongGeniusInfo(songID int64, geniusID int, album, thumbnailURL string) error {
	_, err := db.DB.Exec(`
		UPDATE songs SET genius_id = ?, album = ?, thumbnail_url = ? WHERE id = ?
	`, geniusID, album, thumbnailURL, songID)
	return err
}

// GetSongByAudioPath retrieves a song by its audio path
func GetSongByAudioPath(audioPath string) (*models.Song, error) {
	s := &models.Song{}
	err := db.DB.QueryRow(`
		SELECT id, COALESCE(youtube_id, ''), genius_id, title, artist, COALESCE(album, ''),
		       difficulty, COALESCE(duration_seconds, 0), COALESCE(thumbnail_url, ''),
		       COALESCE(audio_path, ''), created_at
		FROM songs
		WHERE audio_path = ?
	`, audioPath).Scan(&s.ID, &s.YouTubeID, &s.GeniusID, &s.Title, &s.Artist, &s.Album,
		&s.Difficulty, &s.DurationSeconds, &s.ThumbnailURL, &s.AudioPath, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetUnlinkedSongVocab returns song vocabulary that doesn't have a card yet
func GetUnlinkedSongVocab(songID int64) ([]models.SongVocab, error) {
	rows, err := db.DB.Query(`
		SELECT id, song_id, COALESCE(card_id, 0), word, translation, is_key_vocab
		FROM song_vocabulary
		WHERE song_id = ? AND (card_id IS NULL OR card_id = 0)
	`, songID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vocabs []models.SongVocab
	for rows.Next() {
		var v models.SongVocab
		var isKey int
		if err := rows.Scan(&v.ID, &v.SongID, &v.CardID, &v.Word, &v.Translation, &isKey); err != nil {
			return nil, err
		}
		v.IsKeyVocab = isKey == 1
		vocabs = append(vocabs, v)
	}
	return vocabs, rows.Err()
}
