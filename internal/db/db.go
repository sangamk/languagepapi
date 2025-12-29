package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(path string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term TEXT NOT NULL,
		translation TEXT NOT NULL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_reviewed DATETIME,
		review_count INTEGER DEFAULT 0,
		ease_factor REAL DEFAULT 2.5
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date DATE NOT NULL,
		words_reviewed INTEGER DEFAULT 0,
		correct INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_words_last_reviewed ON words(last_reviewed);
	`

	_, err = DB.Exec(schema)
	return err
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
