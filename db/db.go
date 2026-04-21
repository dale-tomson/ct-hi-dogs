package db

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string, seedFile string) {
	var err error

	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("failed to open database", "path", dbPath, "error", err)
		os.Exit(1)
	}

	DB.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA foreign_keys = ON",    // enforce ON DELETE CASCADE
		"PRAGMA journal_mode = WAL",   // allow concurrent reads during writes
		"PRAGMA busy_timeout = 5000",  // wait up to 5s on lock instead of failing
		"PRAGMA synchronous = NORMAL", // safe with WAL, faster than FULL
	}
	for _, p := range pragmas {
		if _, err := DB.Exec(p); err != nil {
			slog.Error("failed to apply pragma", "pragma", p, "error", err)
			os.Exit(1)
		}
	}

	schema := `
CREATE TABLE IF NOT EXISTS breeds (
id   INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT    NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS sub_breeds (
id       INTEGER PRIMARY KEY AUTOINCREMENT,
breed_id INTEGER NOT NULL REFERENCES breeds(id) ON DELETE CASCADE,
name     TEXT    NOT NULL,
UNIQUE(breed_id, name)
);`

	if _, err := DB.Exec(schema); err != nil {
		slog.Error("failed to create schema", "error", err)
		os.Exit(1)
	}

	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM breeds").Scan(&count); err != nil {
		slog.Error("failed to count breeds", "error", err)
		os.Exit(1)
	}
	if count == 0 && seedFile != "" {
		if err := seed(seedFile); err != nil {
			slog.Error("seeding failed", "error", err)
			os.Exit(1)
		}
	}
}

func seed(seedFile string) error {
	data, err := os.ReadFile(seedFile)
	if err != nil {
		slog.Warn("seed file not found, skipping", "file", seedFile)
		return nil // not fatal
	}

	var raw map[string][]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	insertBreed := `INSERT OR IGNORE INTO breeds (name) VALUES (?)`
	insertSub := `INSERT OR IGNORE INTO sub_breeds (breed_id, name) VALUES (?, ?)`

	for breedName, subs := range raw {
		res, err := tx.Exec(insertBreed, breedName)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		breedID, err := res.LastInsertId()
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		for _, sub := range subs {
			if _, err := tx.Exec(insertSub, breedID, sub); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	slog.Info("database seeded", "breeds", len(raw))
	return nil
}
