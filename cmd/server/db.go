package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
}

func InitDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{DB: sqlx.NewDb(db, "sqlite3")}, nil
}

func runMigrations(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT MAX(version) FROM schema_migrations")
	var currentVersion int
	row.Scan(&currentVersion)

	migrationsDir := "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasPrefix(filename, "0") {
			continue
		}

		parts := strings.Split(strings.Split(filename, "_")[0], ".")
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if version <= currentVersion {
			continue
		}

		migrationPath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		log.Printf("Applied migration: %s", filename)
	}

	return tx.Commit()
}
