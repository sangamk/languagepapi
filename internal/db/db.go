package db

import (
	"database/sql"
	_ "embed"

	"languagepapi/internal/db/migrations"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

var DB *sql.DB

func Init(path string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// Enable foreign keys
	if _, err = DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	// Execute schema (creates tables, indexes, and seed data)
	if _, err = DB.Exec(schema); err != nil {
		return err
	}

	// Run any pending migrations
	return migrations.Run(DB)
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
