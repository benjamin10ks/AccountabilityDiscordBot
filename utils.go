package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
	commitStatus, err := checkDailyCommits(db, userID)
	if err != nil {
		log.Printf("Error checking daily commits: %v", err)
		return
	}

	var messageBuilder strings.Builder
	messageBuilder.WriteString(fmt.Sprintf("Daily commit check for <@%s>:\n", userID))

	totalCommitsToday := 0

	for repo, hasCommit := range commitStatus {
		emoji := "❌"
		if hasCommit {
			emoji = "✅"
			totalCommitsToday++
		}
		messageBuilder.WriteString(fmt.Sprintf("%s %s\n", repo, emoji))
	}

	if totalCommitsToday > 0 {
		messageBuilder.WriteString(fmt.Sprintf("Great job <@%s>! You made %d commits today! Keep it up! 🎉", userID, totalCommitsToday))
	} else {
		messageBuilder.WriteString(fmt.Sprintf("Ur a bum <@%s> get on it 😡", userID))
	}

	sendMessage(dg, ChannelID, messageBuilder.String())
}

func scheduleDailyChecks(db *sql.DB, dg *discordgo.Session) {
	for {
		now := time.Now()
		target := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
		// testing 1 minute
		// target := time.Now().Add(1 * time.Minute)
		if now.After(target) {
			target = target.Add(24 * time.Hour)
		}

		time.Sleep(time.Until(target))

		userIDs, err := getAllRegisteredUserIDs(db)
		if err != nil {
			log.Printf("Error getting registered user IDs: %v", err)
			continue
		}

		for _, userID := range userIDs {
			processUserCommits(db, dg, userID)
		}

	}
}

// TODO: check for a commit and add to count of tracked repos commited to for the day
func checkDailyCommits(db *sql.DB, userID string) (map[string]bool, error) {
	repos, err := getReposByUserID(db, userID)
	if err != nil {
		log.Printf("Error getting repo by user ID: %v", err)
		return nil, err
	}

	commitStatus := make(map[string]bool)
	since := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	for _, repo := range repos {
		repoKey := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
		URL := fmt.Sprintf("https://api.github.com/repos/%s/commits?since=%s&per_page=1", repoKey, since)
		res, err := http.Get(URL)
		if err != nil {
			return nil, fmt.Errorf("error making http request: %v", err)
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}

		err = res.Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}

		var commits []any
		if err = json.Unmarshal(data, &commits); err != nil {
			return nil, fmt.Errorf("error parsing json: %v", err)
		}

		commitStatus[repoKey] = len(commits) > 0
	}

	return commitStatus, nil
}

// TODO: implement this function to generate a secure random state token for OAuth flow
func generateStateToken() string {
	return ""
}

func exchangeCodeForToken(code string) (string, error) {
	return "", nil
}

// TODO: implement this function to set up GitHub webhooks for the registered repositories
func createWebhook(accessToken, owner, repo, baseURL string) error {
	return nil
}
