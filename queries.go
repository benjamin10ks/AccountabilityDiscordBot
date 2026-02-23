package main

import (
	"database/sql"
	"log"
)

func getAllRegisteredUserIDs(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT discord_user_id FROM repo_registrations")
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func getRepoByUserID(db *sql.DB, userID string) (string, string, error) {
	row := db.QueryRow("SELECT owner, repo_name FROM repo_registrations WHERE discord_user_id = ?", userID)

	var owner, repo string
	err := row.Scan(&owner, &repo)
	if err != nil {
		return "", "", err
	}
	return owner, repo, nil
}

func getUserIDByOwner(db *sql.DB, owner string) (string, error) {
	row := db.QueryRow("SELECT discord_user_id FROM repo_registrations WHERE owner = ?", owner)

	var userID string
	err := row.Scan(&userID)
	if err != nil {
		return "", err
	}
	err = db.Close()
	if err != nil {
		return "", err
	}
	return userID, nil
}
