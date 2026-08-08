package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	log.Default().Println("Starting up oinkbot-go!")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	discordBotToken := os.Getenv("DISCORD_BOT_TOKEN")
	discord, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		log.Fatal("Failed to load discordgo with given bot token")
	}
	defer discord.Close()
	log.Default().Println("Successfully initialized oinkbot-go with token")

	// discord.ChannelMessageSend("795776459997577256", "sup g")

	/* This is a cool pattern
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	*/
}