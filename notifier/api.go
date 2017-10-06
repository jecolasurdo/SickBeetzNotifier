package notifier

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jecolasurdo/sickbeetznotifier/notifier/spotifyauth"
	"github.com/nlopes/slack"
	"github.com/zmb3/spotify"
)

const (
	channelName       = "sickbeats"
	botName           = "New Sick Beats!"
	spotifyPlaylist   = "SickBeetz"
	lastCheckFileName = ".lastcheck"
)

// Notifier constains components necessary for Spotify to talk to Slack
type Notifier struct {
	SlackAPI      *slack.Client
	SpotifyClient *spotify.Client
	Settings      *BaseSettings
}

// New returns a reference to a Notifier
func New(settings *BaseSettings) *Notifier {
	return &Notifier{
		SlackAPI:      slack.New(settings.SlackToken),
		SpotifyClient: spotifyauth.Authorize(),
		Settings:      settings,
	}
}

// Run begins polling spotify for playlist changes and sends messages to slack when new songs are added.
func (n *Notifier) Run() {
	ticker := time.NewTicker(30 * time.Second)
	for now := range ticker.C {
		log.Println("Poll")
		n.checkSpotifyAndPostToSlack(now)
	}
}

func (n *Notifier) checkSpotifyAndPostToSlack(now time.Time) {
	var (
		addedBy *spotify.User
		err     error
	)
	playlistTracksPage := n.getPlaylistTracks()
	var lastCheck *time.Time
	lastCheck, err = getLastCheck()
	if err != nil {
		log.Fatalf("error occured getting lastCheck date: %v", err)
	}
	for _, track := range playlistTracksPage.Tracks {
		if track.AddedAt > lastCheck.Format(spotify.TimestampLayout) {
			addedBy, err = n.SpotifyClient.GetUsersPublicProfile(spotify.ID(track.AddedBy.ID))
			if err != nil {
				log.Fatalf("error retrieving user information%v", err)
			}
			trackURL := fmt.Sprintf("https://open.spotify.com/track/%v", track.Track.ID)
			var audioFeatures []*spotify.AudioFeatures
			audioFeatures, err = n.SpotifyClient.GetAudioFeatures(track.Track.ID)
			if err != nil {
				log.Println(err)
			}
			nifty := ""
			if len(audioFeatures) > 0 {
				nifty = niftyComment(audioFeatures[0])
			}
			msg := fmt.Sprintf("%v just shared a new track!%v\n%v", addedBy.DisplayName, nifty, trackURL)
			params := slack.PostMessageParameters{
				Username:    botName,
				UnfurlLinks: true,
				UnfurlMedia: true,
			}
			_, _, err = n.SlackAPI.PostMessage(channelName, msg, params)
			if err != nil {
				log.Fatalf("error sending message to channel: %v", err)
			}
		}
	}
	err = setLastCheck(now.UTC())
	if err != nil {
		log.Fatalf("error setting last check time: %v", err)
	}
}

func (n *Notifier) getPlaylistTracks() *spotify.PlaylistTrackPage {
	playlistPage, err := n.SpotifyClient.GetPlaylistsForUser(n.Settings.SpotifyUser)
	if err != nil {
		log.Fatalf("Couldn't get playlists for %v: %v", n.Settings.SpotifyUser, err)
	}

	var selectedPlaylist *spotify.SimplePlaylist
	for _, playlist := range playlistPage.Playlists {
		if playlist.Name == spotifyPlaylist {
			selectedPlaylist = &playlist
			break
		}
	}
	if selectedPlaylist == nil {
		log.Fatalf("playlist not found.")
	}

	playlistTracksPage, err := n.SpotifyClient.GetPlaylistTracks(n.Settings.PlaylistOwner, selectedPlaylist.ID)
	if err != nil {
		log.Fatalf("error getting tracks for playlist: %v", err)
	}
	return playlistTracksPage
}

func niftyComment(track *spotify.AudioFeatures) string {
	comments := []string{}

	if track.Danceability >= 0.9 {
		comments = append(comments, "\nGet your dance shoes on!")
	}

	if track.Danceability <= 0.1 {
		comments = append(comments, "\nChill selection. Very chill.")
	}

	if track.Valence >= 0.9 {
		comments = append(comments, "\nSomeone's feeling happy today!")
	}

	if track.Valence <= 0.1 {
		comments = append(comments, "\nOooh, very contemplative...")
	}

	if len(comments) == 0 {
		return ""
	}

	if len(comments) == 1 {
		return comments[0]
	}

	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(comments) - 1)

	return comments[i]
}
