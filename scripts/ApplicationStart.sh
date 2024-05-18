#!/usr/bin/env bash

start_backend() {
    cd /home/ubuntu/song-recognition
    touch back.txt
    go build -tags netgo -ldflags '-s -w' -o app
    nohup ./app > backend.log 2>&1 &
}

start_client() {
    cd /home/ubuntu/song-recognition/client
    npm install
    npm run build
    nohup serve -s build > client.log 2>&1 &
}

start_backend
