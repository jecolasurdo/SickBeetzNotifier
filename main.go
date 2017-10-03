package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nlopes/slack"
	"github.com/zmb3/spotify"
)

// Log into spotify
// Log into slack
// Check playlist
// Anything new?
// If so, post it!

const (
	redirectURI     = "http://localhost:8080/callback"
	channelName     = "tests"
	botName         = "SlackPlaylistNotifier"
	spotifyUser     = "127658174"
	playlistOwner   = "barsae"
	spotifyPlaylist = "SickBeetz"
)

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadCollaborative)
	ch    = make(chan *spotify.spotifyClient)
	state = "abc123"
)

func main() {
	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))

	// Spotify Auth
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	spotifyClient := <-ch

	// If we're here, we're ready to do stuff...

	user, err := spotifyClient.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	playlistPage, err := spotifyClient.GetPlaylistsForUser(spotifyUser)
	if err != nil {
		log.Fatalf("Couldn't get playlists for %v: %v", spotifyUser, err)
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

	playlistTracksPage, err := spotifyClient.GetPlaylistTracks(playlistOwner, selectedPlaylist.ID)
	if err != nil {
		log.Fatalf("error getting tracks for playlist: %v", err)
	}

	ticker := time.NewTicker(time.Minute)
	lastCheck := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	for now := range ticker.C {
		fmt.Println(lastCheck)
		for _, track := range playlistTracksPage.Tracks {
			if trackAdded, _ := time.Parse(spotify.TimestampLayout, track.AddedAt); trackAdded.After(lastCheck) {
				msg := fmt.Sprintf("%v", track.Track.Name)
				params := slack.PostMessageParameters{
					Username: botName,
				}
				_, _, err := slackAPI.PostMessage(channelName, msg, params)
				if err != nil {
					log.Fatalf("error sending message to channel: %v", err)
				}
			}
		}
		lastCheck = now
	}

}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated spotifyClient
	spotifyClient := auth.NewspotifyClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &spotifyClient
}
