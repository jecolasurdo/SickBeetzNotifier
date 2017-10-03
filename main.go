package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nlopes/slack"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const (
	redirectURI          = "http://localhost:8080/callback"
	channelName          = "tests"
	botName              = "New Sick Beats!"
	spotifyPlaylist      = "SickBeetz"
	spotifyTokenFileName = ".spotifytoken"
	lastCheckFileName    = ".lastcheck"
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
	var lastCheck *time.Time
	lastCheck, err = getLastCheck()
	if err != nil {
		log.Fatalf("error occured getting lastCheck date: %v", err)
	}
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
		err = setLastCheck(now.UTC())
		if err != nil {
			log.Fatalf("error setting last check time: %v", err)
		}
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

func retrieveToken() (*oauth2.Token, error) {
	var (
		err             error
		serializedToken []byte
		token           oauth2.Token
	)
	_, err = os.Stat(spotifyTokenFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if serializedToken, err = ioutil.ReadFile(spotifyTokenFileName); err == nil {
		err = json.Unmarshal(serializedToken, &token)
	}

	return &token, err
}

func saveToken(token *oauth2.Token) error {
	var (
		err             error
		f               *os.File
		serializedToken []byte
	)
	if f, err = os.Create(spotifyTokenFileName); err == nil {
		if serializedToken, err = json.Marshal(*token); err == nil {
			_, err = f.Write(serializedToken)
		}
	}
	return err
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
			nowUTC := time.Now().UTC()
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
