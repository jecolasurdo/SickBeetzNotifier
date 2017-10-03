package notifier

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

// Run begins polling spotify for playlist changes and sends messages to slack when new songs are added.
func Run(settings *Settings) {
	slackAPI := slack.New(settings.SlackToken)
	spotifyClient := spotifyauth.Authorize()

	ticker := time.NewTicker(20 * time.Second)
	for now := range ticker.C {
		log.Println("Poll")
		playlistPage, err := spotifyClient.GetPlaylistsForUser(settings.SpotifyUser)
		if err != nil {
			log.Fatalf("Couldn't get playlists for %v: %v", settings.SpotifyUser, err)
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

		playlistTracksPage, err := spotifyClient.GetPlaylistTracks(settings.PlaylistOwner, selectedPlaylist.ID)
		if err != nil {
			log.Fatalf("error getting tracks for playlist: %v", err)
		}

		var lastCheck *time.Time
		lastCheck, err = getLastCheck()
		if err != nil {
			log.Fatalf("error occured getting lastCheck date: %v", err)
		}
		for _, track := range playlistTracksPage.Tracks {
			if track.AddedAt > lastCheck.Format(spotify.TimestampLayout) {
				msg := fmt.Sprintf("Someone just shared a new track!\n%v\nCheck it out! %v", track.Track.Name, settings.PlaylistURI)
				params := slack.PostMessageParameters{
					Username: botName,
				}
				_, _, err := slackAPI.PostMessage(channelName, msg, params)
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
}

func getLastCheck() (*time.Time, error) {
	var (
		err       error
		rawDate   []byte
		lastCheck time.Time
	)
	_, err = os.Stat(lastCheckFileName)
	if err != nil {
		if os.IsNotExist(err) {
			nowUTC := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
			return &nowUTC, nil
		}
		return nil, err
	}

	if rawDate, err = ioutil.ReadFile(lastCheckFileName); err == nil {
		lastCheck, err = time.Parse(spotify.TimestampLayout, string(rawDate))
	}
	return &lastCheck, err
}

func setLastCheck(lastCheckTime time.Time) error {
	formattedTime := []byte(lastCheckTime.Format(spotify.TimestampLayout))
	return ioutil.WriteFile(lastCheckFileName, formattedTime, 0644)
}
