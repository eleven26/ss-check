#!/bin/bash

# allow specifying different destination directory
DIR="${DIR:-"/usr/local/bin"}"

# prepare the download URL
GITHUB_LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/eleven26/ss-check/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
GITHUB_FILE="ss-check_${GITHUB_LATEST_VERSION//v/}.tar.gz"
GITHUB_URL="https://github.com/eleven26/ss-check/releases/download/${GITHUB_LATEST_VERSION}/${GITHUB_FILE}"

# install/update the local binary
curl -L -o ss-check.tar.gz $GITHUB_URL
tar xzvf ss-check.tar.gz ss-check
sudo mv -f ss-check "$DIR"
rm ss-check.tar.gz
