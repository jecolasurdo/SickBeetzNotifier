FROM golang:1.9

# go's env vars
ENV GOBIN=$GOROOT/bin \
    GOPATH=/build

RUN go get -u github.com/golang/dep/cmd/dep

# SBNs env vars
ENV SBN_SLACK_BOT_NAME="" \
    SBN_SLACK_CHANNEL_NAME="" \
    SBN_SLACK_TOKEN="" \
    SBN_SPOTIFY_PLAYLIST_NAME-="" \
    SBN_SPOTIFY_PLAYLIST_OWNER="" \
    SBN_SPOTIFY_PLAYLIST_URI="" \
    SBN_SPOTIFY_REDIRECT_URI="" \
    SBN_SPOTIFY_USER="" \
    SPOTIFY_ID="" \
    SPOTIFY_SECRET=""

RUN mkdir -p /build/src/github.com/jecolasurdo/sickbeetznotifier
WORKDIR /build/src/github.com/jecolasurdo/sickbeetznotifier
COPY . /build/src/github.com/jecolasurdo/sickbeetznotifier

RUN dep ensure && go build main.go

# run ./main via cron