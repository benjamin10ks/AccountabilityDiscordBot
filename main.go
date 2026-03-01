package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	BotToken       = os.Getenv("DISCORD_BOT_TOKEN")
	GithubClientID = os.Getenv("GITHUB_CLIENT_ID")
	GithubSecret   = os.Getenv("GITHUB_CLIENT_SECRET")
	BaseURL        = os.Getenv("BASE_URL")
)

type PendingAuth struct {
	DiscordUserID string
	Owner         string
	Repo          string
	ExpiresAt     time.Time
}

var (
	pendingAuths   = make(map[string]PendingAuth)
	pendingAuthsMu sync.Mutex
)

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

	err = runMigrations(db)
	if err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	registerCommands(dg, db)

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
		handleGithubCallback(db, w, r)
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

	go scheduleDailyChecks(db, dg)

	log.Println("Bot is now running.")

	select {}
}
