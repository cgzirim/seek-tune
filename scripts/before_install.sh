#!/usr/bin/env bash

sudo apt-get -y update

# install golang
sudo apt-get -y install golang-go

# install nodeJS and npm
sudo apt -y install nodejs
sudo apt -y install npm

# install ffmpeg
sudo apt-get -y install ffmpeg

# install Certbot
DOMAIN="localport.online"
EMAIL="cgzirim@gmail.com"
CERT_DIR="/etc/letsencrypt/live/$DOMAIN"

if [ ! -f "$CERT_DIR" ]; then
    sudo apt install -y certbot
    sudo certbot certonly --standalone -d $DOMAIN --email $EMAIL --agree-tos --non-interactive
    if [ $? -eq 0 ]; then
        sudo apt-get -y install acl
        sudo setfacl -m u:ubuntu:--x /etc/letsencrypt/archive
  fi
fi

# Install MongoDB only if not already present
if [ ! -f "/usr/bin/mongod" ]; then
    sudo apt-get install gnupg curl
    curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | \
       sudo gpg -o /usr/share/keyrings/mongodb-server-7.0.gpg \
       --dearmor
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list
    sudo apt-get update
    sudo apt-get install -y mongodb-org
    sudo apt-get install -y mongosh
fi

sudo rm -rf /home/ubuntu/song-recognition
