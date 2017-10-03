package notifier

import (
	"fmt"
	"log"
	"time"

	"github.com/jecolasurdo/sickbeetznotifier/notifier/spotifyauth"
	"github.com/nlopes/slack"
	"github.com/zmb3/spotify"
)

const (
	channelName       = "tests"
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
	ticker := time.NewTicker(20 * time.Second)
	for now := range ticker.C {
		log.Println("Poll")
		n.checkSpotifyAndPostToSlack(now)
	}
}

func (n *Notifier) checkSpotifyAndPostToSlack(now time.Time) {
	var err error
	playlistTracksPage := n.getPlaylistTracks()
	var lastCheck *time.Time
	lastCheck, err = getLastCheck()
	if err != nil {
		log.Fatalf("error occured getting lastCheck date: %v", err)
	}
	for _, track := range playlistTracksPage.Tracks {
		if track.AddedAt > lastCheck.Format(spotify.TimestampLayout) {
			msg := fmt.Sprintf("Someone just shared a new track!\n%v\nCheck it out! %v", track.Track.Name, n.Settings.PlaylistURI)
			params := slack.PostMessageParameters{
				Username: botName,
			}
			_, _, err := n.SlackAPI.PostMessage(channelName, msg, params)
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
