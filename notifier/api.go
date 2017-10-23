package notifier

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nlopes/slack"
	"github.com/zmb3/spotify"
)

// Notifier constains components necessary for Spotify to talk to Slack
type Notifier struct {
	SlackAPI      *slack.Client
	SpotifyClient *spotify.Client
	Settings      *Settings
}

// New returns a reference to a Notifier
func New(settings *Settings) *Notifier {
	return &Notifier{
		SlackAPI:      slack.New(settings.SlackToken),
		SpotifyClient: Authorize(settings),
		Settings:      settings,
	}
}

// Run begins polling spotify for playlist changes and sends messages to slack when new songs are added.
func (n *Notifier) Run() {
	n.checkSpotifyAndPostToSlack(time.Now().UTC())
}

func (n *Notifier) checkSpotifyAndPostToSlack(now time.Time) {
	var (
		addedBy *spotify.User
		err     error
	)
	playlistTracksPage := n.getPlaylistTracks()
	var lastCheck *time.Time
	lastCheck, err = getLastCheck(n.Settings)
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
				Username:    n.Settings.SlackBotName,
				UnfurlLinks: true,
				UnfurlMedia: true,
			}
			_, _, err = n.SlackAPI.PostMessage(n.Settings.SlackChannelName, msg, params)
			if err != nil {
				log.Fatalf("error sending message to channel: %v", err)
			}
		}
	}
	err = setLastCheck(now.UTC(), n.Settings)
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
		if playlist.Name == n.Settings.SpotifyPlaylistName {
			selectedPlaylist = &playlist
			break
		}
	}
	if selectedPlaylist == nil {
		log.Fatalf("playlist not found.")
	}

	playlistTracksPage, err := n.SpotifyClient.GetPlaylistTracks(n.Settings.SpotifyPlaylistOwner, selectedPlaylist.ID)
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
