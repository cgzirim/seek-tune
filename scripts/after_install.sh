#!/usr/bin/env bash

sudo chown -R ubuntu:ubuntu /home/ubuntu/song-recognition

sudo systemctl start mongod.service
sudo systemctl enable mongod.service

# Generate SSL Cert
cd /home/ubuntu
PUB_IP_ADDRESS=$(curl -s ifconfig.me)
openssl genpkey -algorithm RSA -out song_rec-server.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key song_rec-server.key -out server.csr -subj "/CN=$PUB_IP_ADDRESS"
openssl x509 -req -days 365 -in server.csr -signkey song_rec-server.key -out song_rec-server.crt -extensions v3_req -extfile <(printf "[v3_req]\nsubjectAltName=IP:$PUB_IP_ADDRESS")

sudo rm server.csr

export CERT_PATH=/home/ubuntu/song_rec-server.crt
export CERT_KEY_PATH=/home/ubuntu/song_rec-server.key