#!/bin/sh

# get timestamp of release
NOW=$(date +"%d/%m/%Y %H:%M:%S")

# store release details in case we need to manually restart the container later
echo "RELEASE_TAG=$2 RELEASE_TIMESTAMP=\"$NOW\"" > latest_release.txt

# restart docker container with new image version
cd $1;
RELEASE_TAG=$2 RELEASE_TIMESTAMP=$NOW docker-compose up -d --build;

# prune all non-running images
docker image prune -af
