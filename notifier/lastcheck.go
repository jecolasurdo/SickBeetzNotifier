package notifier

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/zmb3/spotify"
)

func getLastCheck() (*time.Time, error) {
	var (
		err       error
		rawDate   []byte
		lastCheck time.Time
	)
	_, err = os.Stat(lastCheckFileName)
	if err != nil {
		if os.IsNotExist(err) {
			nowUTC := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
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
