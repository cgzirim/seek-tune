#!/usr/bin/env bash

start_backend() {
    cd /home/ubuntu/song-recognition
    go build -tags netgo -ldflags '-s -w' -o app
    nohup ./app > backend.log 2>&1 &
    echo "Backend started successfully."
}

start_client() {
    cd /home/ubuntu/song-recognition/client
    nvm use 16
    npm run build
    nohup serve -s build > client.log 2>&1 &
    echo "Client started successfully."
}

start_backend && start_client
