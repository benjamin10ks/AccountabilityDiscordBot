package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func registerCommands(dg *discordgo.Session, db *sql.DB) {
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

			owner, repo := parts[0], parts[1]

			stateToken := generateStateToken()

			pendingAuthsMu.Lock()
			pendingAuths[stateToken] = PendingAuth{
				DiscordUserID: userID,
				Owner:         owner,
				Repo:          repo,
				ExpiresAt:     time.Now().Add(10 * time.Minute),
			}
			pendingAuthsMu.Unlock()

			authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&scope=admin:repo_hook&state=%s", GithubClientID, stateToken)

			err := registerRepo(db, userID, owner, repo)
			if err != nil {
				log.Printf("Error registering repo: %v", err)
			}

			log.Printf("Registering repo '%s' for user %s", repoInput, userID)

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Click here to authorize Github access %s\n*(Link expires in 10 minutes)", authURL),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Printf("Error responding to interaction: %v", err)
			}
		}
	})
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "register",
		Description: "Register a GitHub repository to watch",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "repo",
				Description: "Repository in format owner/repo",
				Required:    true,
			},
		},
	},
}
