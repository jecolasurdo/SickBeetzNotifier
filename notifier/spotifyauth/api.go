package spotifyauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const (
	redirectURI          = "http://localhost:8080/callback"
	spotifyTokenFileName = ".spotifytoken"
)

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadCollaborative)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

// Authorize initializes authorization for Spotify.
//
// First attempts to find an existing token locally.
// If no token is found locally, the user is prompted to authorize at a URL.
// The method blocks until authorization is complete.
func Authorize() *spotify.Client {
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
