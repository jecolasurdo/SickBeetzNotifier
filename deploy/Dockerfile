FROM golang:1.9


# go's env vars
ENV GOBIN=$GOROOT/bin \
    GOPATH=/build

RUN go get -u github.com/golang/dep/cmd/dep

RUN apt-get update && apt-get -y install cron

# SBNs env vars
ENV SBN_FILE_PATH="/bin/sbn/" \
    SBN_SLACK_BOT_NAME="" \
    SBN_SLACK_CHANNEL_NAME="" \
    SBN_SLACK_TOKEN="" \
    SBN_SPOTIFY_PLAYLIST_NAME-="" \
    SBN_SPOTIFY_PLAYLIST_OWNER="" \
    SBN_SPOTIFY_PLAYLIST_URI="" \
    SBN_SPOTIFY_REDIRECT_URI="" \
    SBN_SPOTIFY_USER="" \
    SPOTIFY_ID="" \
    SPOTIFY_SECRET=""

RUN mkdir /bin/sbn
RUN mkdir -p /build/src/github.com/jecolasurdo/sickbeetznotifier
WORKDIR /build/src/github.com/jecolasurdo/sickbeetznotifier
COPY . /build/src/github.com/jecolasurdo/sickbeetznotifier
RUN mv /build/src/github.com/jecolasurdo/sickbeetznotifier/.lastcheck /bin/sbn/.lastcheck
RUN dep ensure && go install 

EXPOSE 8080

RUN (crontab -l 2> /dev/null ; echo "*/1 * * * * sickbeetznotifier >> /bin/sbn/log 2>> /bin/sbn/log") | crontab -

# Passing env vars to cron: https://stackoverflow.com/a/41938139/3434541
CMD printenv >> /etc/environment && cron -f