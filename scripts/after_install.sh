#!/usr/bin/env bash

sudo chown -R ubuntu:ubuntu /home/ubuntu/song-recognition

sudo systemctl start mongod.service
sudo systemctl enable mongod.service