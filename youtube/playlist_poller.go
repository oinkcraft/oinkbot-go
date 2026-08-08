package youtube

import (
	"context"
	"fmt"
	"log"

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
func Setup(ytApiKey string, discord *discordgo.Session, ytConfig *Config) {
	ytClient, err := youtube.NewService(context.TODO(), option.WithAPIKey(ytApiKey))
	if err != nil {
		log.Fatal("Failed to load youtube with the given api key!")
	}
	YTClient = ytClient
	YTConfig = ytConfig
	Discord = discord
	StartPollingPlaylists()
}

// Poll the playlists every 30 minutes
func StartPollingPlaylists() {
	cowdinoChannelId := YTConfig.CowdinoChannelId
	fracThinkingPlaylistId := YTConfig.FracturedThinkingPlaylist
	fracThinkingRoleId := YTConfig.FracturedThinkingRoleId

	// Our approach is going to be to get the latest video, then compare what playlist it's on
	// I can also just check for `(FT [09])` regex, but that relies on me always remembering to do that lol
	latestVidResp, err := YTClient.Search.
		List([]string{"snippet"}).
		ChannelId(cowdinoChannelId).
		Type("video").
		Order("date").
		MaxResults(1).
		Do()
	if err != nil {
		log.Default().Println("Failed to get latest video. Err: ", err)
	}
	latestVideo := latestVidResp.Items[0]
	latestVideoId := latestVideo.Id.VideoId

	// PICKUP: call func to write to a file that this was the last one sent etc.
	//		check video id first before calling playlist
	alreadySent := false

	if !alreadySent {
		checkIsInPlaylistResp, err := YTClient.PlaylistItems.
			List([]string{"id"}).
			PlaylistId(fracThinkingPlaylistId).
			VideoId(latestVideoId).
			MaxResults(1).
			Do()
		if err != nil {
			log.Default().Println("Failed to get a result from playlist. Err: ", err)
		}
		if len(checkIsInPlaylistResp.Items) > 0 {
			vidUrlStr := "https://www.youtube.com/watch?v=" + latestVideoId
			msgStr := fmt.Sprintf("<@&%s> **NEW FRACTURED THINKING**\n\n%s", fracThinkingRoleId, vidUrlStr)
			
			Discord.ChannelMessageSend(YTConfig.DiscordNotifChannelId, msgStr)
		}
	}


	// TODO: The other two
	// NOTE: There's another way to go about latest video published by someone

	// TODO: Re-run this function after 30 minutes elapses
}