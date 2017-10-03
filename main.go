package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nlopes/slack"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// Log into spotify
// Log into slack
// Check playlist
// Anything new?
// If so, post it!

const (
	redirectURI     = "http://localhost:8080/callback"
	channelName     = "tests"
	botName         = "New Sick Beats!"
	spotifyPlaylist = "SickBeetz"
)

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadCollaborative)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

// main is the entrypoint for the program
//
// Will panic if expected environment variables are not set (see settings.go)
func main() {
	s := initializeSettings()
	slackAPI := slack.New(s.SlackToken)
	spotifyClient := startSpotifyAuthorization()

	playlistPage, err := spotifyClient.GetPlaylistsForUser(s.SpotifyUser)
	if err != nil {
		log.Fatalf("Couldn't get playlists for %v: %v", s.SpotifyUser, err)
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

	playlistTracksPage, err := spotifyClient.GetPlaylistTracks(s.PlaylistOwner, selectedPlaylist.ID)
	if err != nil {
		log.Fatalf("error getting tracks for playlist: %v", err)
	}

	ticker := time.NewTicker(60 * time.Second)
	lastCheck := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	for now := range ticker.C {
		for _, track := range playlistTracksPage.Tracks {
			trackAdded, _ := time.Parse(spotify.TimestampLayout, track.AddedAt)
			fmt.Printf("Track: %v\nAdded: %v\nLast Check: %v\n-----------\n", track.Track.Name, trackAdded, lastCheck)
			if track.AddedAt > lastCheck.Format(spotify.TimestampLayout) {
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
		lastCheck = now.UTC()
		fmt.Println("-----------------------------------------")
	}
}

func startSpotifyAuthorization() *spotify.Client {
	token, err := retrieveToken()
	if err != nil {
		log.Fatalf("error checking for existing token: %v", err)
	}
	if token != nil {
		client := auth.NewClient(token)
		log.Println("Found existing token!")
		return &client
	}

	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	return <-ch
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
	spotifyClient := auth.NewClient(tok)

	err = saveToken(tok)
	if err != nil {
		log.Fatalf("error trying to save token for later use:, %v", err)
	}

	fmt.Fprintf(w, "Login Completed!")
	ch <- &spotifyClient
}

// retrieveToken is not implmented, but it would be nice if the token was persisted so we don't have to do
// an OAUTH http request every time the program is initiated.
// This authorization shortcoming is why the program currently has it's own internal ticker
// rather than just relying on an external scheduler, which would be preferrable.
func retrieveToken() (*oauth2.Token, error) {
	return nil, nil
}

// saveToken is not implemented. See notes about retrieveToken.
func saveToken(token *oauth2.Token) error {
	return nil
}
