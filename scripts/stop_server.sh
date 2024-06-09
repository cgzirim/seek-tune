#!/usr/bin/env bash

HTTP_PID=$(sudo lsof -t -i:5000)
HTTPS_PID=$(sudo lsof -t -i:4443)


if [ -n "$HTTP_PID" ]; then
  sudo kill -9 $HTTP_PID
fi

if [ -n "$HTTPS_PID" ]; then
  sudo kill -9 $HTTPS_PID
fi
