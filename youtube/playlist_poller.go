package youtube

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

// Responsible for polling whatever playlists we care about for respective videos
// TODO: Make these a struct and just pull from that
var (
	YTClient *youtube.Service
	YTConfig *Config
	Discord *discordgo.Session
	SQLite *sql.DB
)

// TODO: Break these down more haha, but whatever for now
type Config struct {
	CowdinoChannelId string `json:"cowdino_channel_id"`
	FracturedThinkingPlaylist string `json:"fractured_thinking_playlist"`
	DevlogPlaylist string `json:"devlog_playlist"`
	LifestylePlaylist string `json:"lifestyle_playlist"`
	CowdinoPlaylist string `json:"cowdino_playlist"`
	FracturedThinkingRoleId string `json:"fractured_thinking_role_id"`
	DiscordNotifChannelId string `json:"discord_notif_channel_id"`
}

// NOTE: not using init() here because ordering
// Start polling youtube and channels for a respective video
func Setup(ytApiKey string, discord *discordgo.Session, db *sql.DB, ytConfig *Config) {
	ytClient, err := youtube.NewService(context.TODO(), option.WithAPIKey(ytApiKey))
	if err != nil {
		log.Fatal("Failed to load youtube with the given api key!")
	}
	YTClient = ytClient
	YTConfig = ytConfig
	Discord = discord
	SQLite = db

	InitDB()
	StartPollingPlaylists()
}

// Poll the playlists every 30 minutes
func StartPollingPlaylists() {
	PollAllChannels()

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		PollAllChannels()
	}
}

// Split out for the recursive call
func PollAllChannels() {
	cowdinoChannelId := YTConfig.CowdinoChannelId
	fracThinkingPlaylistId := YTConfig.FracturedThinkingPlaylist
	fracThinkingRoleId := YTConfig.FracturedThinkingRoleId

	PollLatestVideoFor(cowdinoChannelId, fracThinkingPlaylistId, fracThinkingRoleId, "🫯 NEW FRACTURED THINKING")
}

func PollLatestVideoFor(channelId, playlistId, roleId, customMessage string) {
	latestVidResp, err := YTClient.Search.
		List([]string{"snippet"}).
		ChannelId(channelId).
		Type("video").
		Order("date").
		MaxResults(1).
		Do()
	if err != nil {
		log.Default().Println("Failed to get latest video. Err: ", err)
		return
	}

	if len(latestVidResp.Items) == 0 {
		return
	}

	latestVideoId := latestVidResp.Items[0].Id.VideoId
	latestVideoTitle := latestVidResp.Items[0].Snippet.Title

	if IsVideoInDb(latestVideoId) {
		log.Default().Printf("Video id %s already in db. Skipping\n", latestVideoId)
		return
	}

	checkIsInPlaylistResp, err := YTClient.PlaylistItems.
		List([]string{"id"}).
		PlaylistId(playlistId).
		VideoId(latestVideoId).
		MaxResults(1).
		Do()
	if err != nil {
		log.Default().Println("Failed to get a result from playlist. Err: ", err)
		return
	}

	if len(checkIsInPlaylistResp.Items) == 0 {
		return
	}

	vidUrlStr := "https://www.youtube.com/watch?v=" + latestVideoId
	msgStr := fmt.Sprintf(
		"## %s\n\n<@&%s>\n%s",
		customMessage,
		roleId,
		vidUrlStr,
	)

	if _, err := Discord.ChannelMessageSend(
		YTConfig.DiscordNotifChannelId,
		msgStr,
	); err != nil {
		log.Default().Println("Failed to send video to Discord. Err: ", err)
		return
	}

	log.Default().Printf("%s - Video sent to channel on Discord\n", latestVideoId)

	WriteVideoToDb(
		latestVideoId,
		latestVideoTitle,
		playlistId,
	)
}

// We don't really need all the historical data, but eh why not. This table isn't going to get big any time soon
func IsVideoInDb(videoId string) (found bool) {
	err := SQLite.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM videos WHERE video_id = ?)
	`, videoId).Scan(&found)

	if err != nil {
		log.Default().Println("Failed to run a check if a video exists! Returning true to prevent spam. Error: ", err)
		found = true
	}

	return
}

func WriteVideoToDb(videoId, title, playlistId string) {
	_, err := SQLite.Exec(
		`INSERT INTO videos (video_id, title, playlist_id, discovered_at)
		 VALUES(?, ?, ?, ?)`,
		videoId, title, playlistId, time.Now())
	if err != nil {
		log.Default().Printf("Failed to write video %s to db!\n", videoId)
	} else {
		log.Default().Printf("%s - Video ID written to db\n", videoId)
	}
}

// Set up DB table, keep block text out of setup func
func InitDB() {
	_, err := SQLite.Exec(`
		CREATE TABLE IF NOT EXISTS videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			video_id TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			playlist_id TEXT NOT NULL,
			discovered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)

	if err != nil {
		log.Fatal("Failed to initialize sqlite table in youtube package")
	}
}