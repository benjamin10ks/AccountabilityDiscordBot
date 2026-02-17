package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

var (
	BotToken  = os.Getenv("DISCORD_BOT_TOKEN")
	ChannelID = os.Getenv("DISCORD_CHANNEL_ID")
)

func main() {
	dg, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")

	select {}
}
