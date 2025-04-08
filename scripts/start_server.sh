#!/usr/bin/env bash

start_server() {
    cd /home/ubuntu/song-recognition/server

    export SERVE_HTTPS="true"
    export CERT_KEY="/etc/letsencrypt/live/localport.online/privkey.pem"
    export CERT_FILE="/etc/letsencrypt/live/localport.online/fullchain.pem"

    go build -tags netgo -ldflags '-s -w' -o app
    sudo setcap CAP_NET_BIND_SERVICE+ep app
    nohup ./app serve -proto https -p 4443 > backend.log 2>&1 &
}

start_client() {
    cd /home/ubuntu/song-recognition/client
    npm install
    npm run build
    nohup serve -s build > client.log 2>&1 &
}

start_server
