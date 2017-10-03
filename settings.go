package main

import "os"

type settings struct {
	SlackToken    string
	SpotifyID     string
	SpotifySecret string
	SpotifyUser   string
	PlaylistOwner string
}

func initializeSettings() *settings {
	s := settings{
		PlaylistOwner: os.Getenv("PLAYLIST_OWNER"),
		SlackToken:    os.Getenv("SLACK_TOKEN"),
		SpotifyID:     os.Getenv("SPOTIFY_ID"),
		SpotifySecret: os.Getenv("SPOTIFY_SECRET"),
		SpotifyUser:   os.Getenv("SPOTIFY_USER"),
	}
	if s.PlaylistOwner == "" ||
		s.SlackToken == "" ||
		s.SpotifyID == "" ||
		s.SpotifySecret == "" ||
		s.SpotifyUser == "" {
		panic("One or more expected environment variables not set.")
	}
	return &s
}
