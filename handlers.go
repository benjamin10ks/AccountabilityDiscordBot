package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

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
func handleGithubCallback(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	pendingAuthsMu.Lock()
	pending, ok := pendingAuths[state]
	delete(pendingAuths, state)
	pendingAuthsMu.Unlock()

	if !ok || time.Now().After(pending.ExpiresAt) {
		http.Error(w, "Invalid or expired state parameter", http.StatusBadRequest)
		return
	}

	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Error exchanging code for token", http.StatusInternalServerError)
		return
	}

	err = storesGithubToken(db, pending.DiscordUserID, accessToken)
	if err != nil {
		http.Error(w, "Error storing GitHub token", http.StatusInternalServerError)
		return
	}
	webhookURL := fmt.Sprintf("%s/webhook", BaseURL)
	err = createWebhook(accessToken, pending.Owner, pending.Repo, webhookURL)
	if err != nil {
		http.Error(w, "Error creating GitHub webhook", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully authenticated user %s for repo %s/%s", pending.DiscordUserID, pending.Owner, pending.Repo)
}
