#!/usr/bin/env bash

start_backend() {
    cd /home/ubuntu/song-recognition
    touch back.txt
    go build -tags netgo -ldflags '-s -w' -o app
    nohup ./app > backend.log 2>&1 &
}

start_client() {
    cd /home/ubuntu/song-recognition/client
    touch client.txt

    export NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    
    nvm install 16
    nvm use 16
    npm install
    npm run build
    nohup serve -s build > client.log 2>&1 &
}

start_backend && start_client
