package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	BotToken       = os.Getenv("DISCORD_BOT_TOKEN")
	ChannelID      = os.Getenv("DISCORD_CHANNEL_ID")
	GithubClientID = os.Getenv("GITHUB_CLIENT_ID")
	GithubSecret   = os.Getenv("GITHUB_CLIENT_SECRET")
	BaseURL        = os.Getenv("BASE_URL")
)

type PushPayload struct {
	Commits []struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commits"`
}

type CommitResponse []struct {
	Commit struct {
		Message string `json:"message"`
	} `json:"commit"`
}

func main() {
	dg, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	db, err := sql.Open("sqlite3", "./bot.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	runMigrations(db)

	// Registers commands
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		switch i.ApplicationCommandData().Name {
		case "register":
			repoInput := i.ApplicationCommandData().Options[0].StringValue()
			userID := i.Member.User.ID

			parts := strings.Split(repoInput, "/")
			if len(parts) != 2 {
				log.Printf("Invalid repo format: %s", repoInput)
				return
			}

			owner := parts[0]
			repo := parts[1]

			err := registerRepo(db, userID, owner, repo)
			if err != nil {
				log.Printf("Error registering repo: %v", err)
			}

			log.Printf("Registering repo '%s' for user %s", repoInput, userID)

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("<@%s> Now watching %s", userID, repoInput),
				},
			})
			if err != nil {
				log.Printf("Error responding to interaction: %v", err)
			}
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer func() {
		err := dg.Close()
		if err != nil {
			log.Printf("Error closing Discord session: %v", err)
		}
	}()

	appID := dg.State.User.ID

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(appID, "", cmd)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", cmd.Name, err)
		}
	}

	http.HandleFunc("/github/callback", func(w http.ResponseWriter, r *http.Request) {
		handleGithubCallback(db, dg, w, r)
	})

	go func() {
		http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
			handleWebhook(db, dg, w, r)
		})
		log.Println("Starting webhook server on :8080")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	go func() {
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
	}()

	log.Println("Bot is now running.")

	select {}
}
