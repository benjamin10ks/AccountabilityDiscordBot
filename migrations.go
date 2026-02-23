package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func runMigrations(db *sql.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS repo_registrations (
		discord_user_id TEXT PRIMARY KEY,
		owner TEXT NOT NULL,
		repo_name TEXT NOT NULL
		);`

	_, err := db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
}
