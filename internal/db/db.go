package db

import (
	"database/sql"
	_ "embed"

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
	_, err = DB.Exec(schema)
	return err
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
