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

func sendMessage(dg *discordgo.Session, channelID, message string) {
	_, err := dg.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
	log.Printf("Sent message: %s", message)
}

func processUserCommits(db *sql.DB, dg *discordgo.Session, userID string) {
	commits, err := checkDailyCommits(db, userID)
	if err != nil {
		log.Printf("Error checking daily commits: %v", err)
		return
	}
	if len(*commits) > 0 {
		sendMessage(dg, ChannelID, fmt.Sprintf("<@%s> Daily commit check: %d commits found for today!", userID, len(*commits)))
	} else {
		sendMessage(dg, ChannelID, fmt.Sprintf("Ur a bum get on it <@%s>", userID))
	}
}

func checkDailyCommits(db *sql.DB, userID string) (*CommitResponse, error) {
	owner, repo, err := getRepoByUserID(db, userID)
	if err != nil {
		log.Printf("Error getting repo by user ID: %v", err)
	}

	since := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	URL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?since=%s", owner, repo, since)
	res, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("error making http request: %v", err)
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var commits CommitResponse
	if err = json.Unmarshal(data, &commits); err != nil {
		return nil, fmt.Errorf("error parsing json: %v", err)
	}

	return &commits, nil
}

// TODO: implement this function to set up GitHub webhooks for the registered repositories
func createWebhook(owner, repo string) error {
	return nil
}
