package notifier

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Settings represents the core settings required for the application to integrate with Slack and Spotify
type Settings struct {
	SlackBotName         string `required:"true" split_words:"true"`
	SlackChannelName     string `required:"true" split_words:"true"`
	SlackToken           string `required:"true" split_words:"true"`
	SpotifyPlaylistName  string `required:"true" split_words:"true"`
	SpotifyPlaylistOwner string `required:"true" split_words:"true"`
	SpotifyPlaylistURI   string `required:"true" split_words:"true"`
	SpotifyRedirectURI   string `required:"true" split_words:"true"`
	SpotifyUser          string `required:"true" split_words:"true"`
}

// InitializeSettingsFromEnvVars initializes settings from environment variables.
// Panics if any variables are not set.
func InitializeSettingsFromEnvVars() *Settings {
	var s Settings
	err := envconfig.Process("sbn", &s)
	if err != nil {
		log.Fatal(err)
	}
	return &s
}
