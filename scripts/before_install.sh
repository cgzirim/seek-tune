#!/usr/bin/env bash

touch /home/ubuntu/status.txt

sudo apt-get -y update
# sudo rm -rf /home/ubuntu/install
echo "A" >> /home/ubuntu/status.txt

if [ ! -f "/home/ubuntu/install" ]; then
    # install CodeDeploy agent
    echo "B" >> /home/ubuntu/status.txt
    sudo apt-get -y install ruby
    sudo apt-get -y install wget
    cd /home/ubuntu
    wget https://aws-codedeploy-eu-north-1.s3.amazonaws.com/latest/install
    sudo chmod +x ./install 
    sudo ./install auto
    sudo apt-get install -y python3-pip
    sudo pip install awscli
    echo "C" >> /home/ubuntu/status.txt
fi

# install golang
sudo apt-get -y install golang-go
 echo "D" >> /home/ubuntu/status.txt

# install nodeJS and npm
sudo apt -y install nodejs
echo "E" >> /home/ubuntu/status.txt
sudo apt -y install npm
echo "F" >> /home/ubuntu/status.txt

# Install MongoDB only if not already present
if [ ! -f "/usr/bin/mongod" ]; then
    echo "G" >> /home/ubuntu/status.txt
    sudo apt-get install gnupg curl
    curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | \
       sudo gpg -o /usr/share/keyrings/mongodb-server-7.0.gpg \
       --dearmor
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list
    sudo apt-get update
    sudo apt-get install -y mongodb-org
    sudo apt-get install -y mongosh
    echo "H" >> /home/ubuntu/status.txt
fi
