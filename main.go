package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	"com.redstoneoinkcraft.oinkbot/m/youtube"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

type Config struct {
	YouTube *youtube.Config `json:"youtube"`
}

func main() {
	log.Default().Println("Starting up oinkbot-go!")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	// Setup the bot stuff
	discordBotToken := os.Getenv("DISCORD_BOT_TOKEN")
	discord, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		log.Fatal("Failed to load discordgo with given bot token")
	}
	defer discord.Close()
	log.Default().Println("Successfully initialized oinkbot-go with token")

	// Set up sqlite db connection
	db, err := sql.Open("sqlite", "oinkbot.db")
	if err != nil {
		log.Fatal("Failed to load the sqlite db")
	}
	defer db.Close()

	// Load configuration
	var config Config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("Failed to load configuration json!")
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatal("Failed to unmarshal config json")
	}

	// Kick off youtube poller
	youtube.Setup(os.Getenv("YOUTUBE_API_KEY"), discord, db, config.YouTube)
	log.Default().Println("Finished youtube setup run")

	// Run until exit signal. Pattern taken from one of the discordgo examples
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Default().Println("Exiting- goodbye!")
}