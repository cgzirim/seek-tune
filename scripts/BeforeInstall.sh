#!/usr/bin/env bash

# clear codedeploy-agent files for a fresh install
# sudo rm -rf /home/ubuntu/install

touch /home/ubuntu/status.txt
sudo apt-get -y update

if [ ! -d "/home/ubuntu/install" ]; then
    # install CodeDeploy agent
    sudo apt-get -y install ruby
    sudo apt-get -y install wget
    cd /home/ubuntu
    wget https://aws-codedeploy-eu-north-1.s3.amazonaws.com/latest/install
    sudo chmod +x ./install 
    sudo ./install auto
else
    echo "CodeDeploy agent already installed."
fi

echo "A" >> /home/ubuntu/status.txt

# install golang
sudo apt-get -y install golang-go
echo "Installed Golang" >> /home/ubuntu/status.txt

# install nodeJS and nvm
sudo apt -y install nodejs
echo "nodeJS installed successfully" >> /home/ubuntu/status.txt
sudo apt -u install npm
echo "npm installed successfully" >> /home/ubuntu/status.txt
wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
echo "nvm installed successfully" >> /home/ubuntu/status.txt

# install MongoDB
sudo apt -y install mongodb
echo "mongodb installed successfully" >> /home/ubuntu/status.txt
