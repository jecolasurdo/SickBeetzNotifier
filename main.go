package main

import (
	"github.com/jecolasurdo/sickbeetznotifier/notifier"
)

func main() {
	s := notifier.InitializeSettingsFromEnvVars()
	n := notifier.New(s)
	n.Run()
}
