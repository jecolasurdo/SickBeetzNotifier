package main

import (
	"log"

	"github.com/jecolasurdo/sickbeetznotifier/notifier"
)

func main() {
	log.Println("Initializing env vars...")
	s := notifier.InitializeSettingsFromEnvVars()
	log.Println("Getting new notifier...")
	n := notifier.New(s)
	log.Println("Running notifier...")
	n.Run()
	log.Println("Done.")
}
