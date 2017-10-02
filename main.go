package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	spotifyPlaylist = "SickBeetz"
)

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadCollaborative)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	// Slack Auth
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	params := slack.PostMessageParameters{
		Username: botName,
	}
	_, _, err := api.PostMessage(channelName, "Hello world", params)
	if err != nil {
		fmt.Println("Error sending message to channel")
		fmt.Println(err)
	}

	// Spotify Auth
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	client := <-ch

	// If we're here, we're ready to do stuff...

	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	playlistPage, err := client.GetPlaylistsForUser(spotifyUser)
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
		log.Fatalf("Playlist not found.")
	}

	fmt.Println(selectedPlaylist.Name)
	fmt.Printf("Number of tracks: %v", selectedPlaylist.Tracks.Total)

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
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}
