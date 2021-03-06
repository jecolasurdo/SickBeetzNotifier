package notifier

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

var (
	auth  spotify.Authenticator
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

// Authorize initializes authorization for Spotify.
//
// First attempts to find an existing token locally.
// If no token is found locally, the user is prompted to authorize at a URL.
// The method blocks until authorization is complete.
func Authorize(s *Settings) *spotify.Client {
	token, err := retrieveToken(s)
	if err != nil {
		log.Fatalf("error checking for existing token: %v", err)
	}
	if token != nil {

		client := auth.NewClient(token)
		log.Println("Found existing token!")
		return &client
	}
	auth = spotify.NewAuthenticator(s.SpotifyRedirectURI, spotify.ScopePlaylistReadCollaborative)
	http.Handle("/callback", newAuthCompleter(s))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	return <-ch
}

func retrieveToken(s *Settings) (*oauth2.Token, error) {
	var (
		err             error
		serializedToken []byte
		token           oauth2.Token
	)
	_, err = os.Stat(s.SpotifyTokenFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if serializedToken, err = ioutil.ReadFile(s.SpotifyTokenFile()); err == nil {
		err = json.Unmarshal(serializedToken, &token)
	}

	return &token, err
}

type authCompleter struct {
	Settings *Settings
}

func newAuthCompleter(settings *Settings) *authCompleter {
	return &authCompleter{
		Settings: settings,
	}
}

func (a *authCompleter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	err = a.saveToken(tok)
	if err != nil {
		log.Fatalf("error trying to save token for later use:, %v", err)
	}

	fmt.Fprintf(w, "Login Completed!")
	ch <- &spotifyClient
}

func (a *authCompleter) saveToken(token *oauth2.Token) error {
	var (
		err             error
		f               *os.File
		serializedToken []byte
	)
	if f, err = os.Create(a.Settings.SpotifyTokenFile()); err == nil {
		if serializedToken, err = json.Marshal(*token); err == nil {
			_, err = f.Write(serializedToken)
		}
	}
	return err
}
