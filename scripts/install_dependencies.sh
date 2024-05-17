#!/usr/bin/env bash

# clear codedeploy-agent files for a fresh install
sudo rm -rf /home/ubuntu/install

# install CodeDeploy agent
sudo apt-get -y update
sudo apt-get -y install ruby
sudo apt-get -y install wget
cd /home/ubuntu
wget https://aws-codedeploy-eu-north-1.s3.amazonaws.com/latest/install
sudo chmod +x ./install 
sudo ./install auto

# install golang
sudo apt-get -y install golang-go

# install nodeJS and nvm
sudo apt install nodejs
sudo apt install npm
wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
