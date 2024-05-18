#!/usr/bin/env bash

set -e  # Exit immediately if a command exits with a non-zero status
set -x  # Print commands and their arguments as they are executed

# Log function to capture error messages
log() {
    echo "$1" >> /var/log/codedeploy_before_install.log
}

log "Updating apt-get"
sudo apt-get -y update || { log "apt-get update failed"; exit 1; }

log "Checking for /home/ubuntu/install"
if [ ! -f "/home/ubuntu/install" ]; then
    log "Installing CodeDeploy agent"
    sudo apt-get -y install ruby || { log "ruby installation failed"; exit 1; }
    sudo apt-get -y install wget || { log "wget installation failed"; exit 1; }
    cd /home/ubuntu
    wget https://aws-codedeploy-eu-north-1.s3.amazonaws.com/latest/install || { log "wget failed"; exit 1; }
    sudo chmod +x ./install
    sudo ./install auto || { log "CodeDeploy agent installation failed"; exit 1; }
fi

log "Installing golang-go"
sudo apt-get -y install golang-go || { log "golang-go installation failed"; exit 1; }

log "Installing nodejs and npm"
sudo apt -y install nodejs || { log "nodejs installation failed"; exit 1; }
sudo apt -y install npm || { log "npm installation failed"; exit 1; }

log "Checking for MongoDB installation"
if [ ! -f "/usr/bin/mongod" ]; then
    log "Installing MongoDB"
    sudo apt-get install -y gnupg curl || { log "gnupg or curl installation failed"; exit 1; }
    curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | sudo gpg -o /usr/share/keyrings/mongodb-server-7.0.gpg --dearmor || { log "MongoDB GPG key download failed"; exit 1; }
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list || { log "MongoDB repository addition failed"; exit 1; }
    sudo apt-get update || { log "apt-get update failed after adding MongoDB repository"; exit 1; }
    sudo apt-get install -y mongodb-org || { log "mongodb-org installation failed"; exit 1; }
    sudo apt-get install -y mongosh || { log "mongosh installation failed"; exit 1; }
fi

log "BeforeInstall script completed successfully"
