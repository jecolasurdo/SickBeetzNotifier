package notifier

import "os"

// Settings represents the core settings required for the application to integrate with Slack and Spotify
type Settings struct {
	SlackToken    string
	SpotifyUser   string
	PlaylistOwner string
	SpotifyID     string
	SpotifySecret string
	PlaylistURI   string
}

// InitializeSettingsFromEnvVars initializes settings from environment variables.
// Panics if any variables are not set.
func InitializeSettingsFromEnvVars() *Settings {
	s := Settings{
		PlaylistOwner: os.Getenv("PLAYLIST_OWNER"),
		SlackToken:    os.Getenv("SLACK_TOKEN"),
		SpotifyUser:   os.Getenv("SPOTIFY_USER"),
		SpotifyID:     os.Getenv("SPOTIFY_ID"),
		SpotifySecret: os.Getenv("SPOTIFY_SECRET"),
		PlaylistURI:   os.Getenv("PLAYLIST_URI"),
	}
	if s.PlaylistOwner == "" ||
		s.SlackToken == "" ||
		s.SpotifyID == "" ||
		s.SpotifySecret == "" ||
		s.SpotifyUser == "" ||
		s.PlaylistURI == "" {
		panic("One or more expected environment variables not set.")
	}
	return &s
}
