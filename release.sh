#!/bin/sh

# get timestamp of release
NOW=$(date +"%Y-%m-%d %H:%M:%S")

# change to project directory
cd $1;

# update shell script with current release data in case we need to manually restart the container later
echo "#\!/bin/sh\n\nRELEASE_TAG=$2 RELEASE_TIMESTAMP=\"$NOW\" docker-compose up -d --build;" > re-release.sh

# restart docker container with new image version
/bin/sh ./re-release.sh

# prune all non-running images
docker image prune -af
