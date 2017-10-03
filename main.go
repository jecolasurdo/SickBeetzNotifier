package main

import "github.com/jecolasurdo/sickbeetznotifier/notifier"

func main() {
	s := notifier.InitializeSettingsFromEnvVars()
	notifier.Run(s)
}
