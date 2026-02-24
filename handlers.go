package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func handleWebhook(db *sql.DB, dg *discordgo.Session, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
	}
	log.Printf("Received webhook: %s", string(body))

	var payload PushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing JSON: %v", err)
	}

	owner := payload.Commits[0].Author.Name

	userID, err := getUserIDByOwner(db, owner)
	if err != nil {
		log.Printf("Error getting user ID by owner: %v", err)
	}

	sendMessage(dg, ChannelID, fmt.Sprintf("<@%s> New commit by %s: %s", userID, owner, payload.Commits[0].Message))
	w.WriteHeader(http.StatusOK)
}

// TODO: finish this function to handle the github callback and exchange the code for an access token, then save the access token in the database
func handleGithubCallback(db *sql.DB, dg *discordgo.Session, w http.ResponseWriter, r *http.Request) {
}
