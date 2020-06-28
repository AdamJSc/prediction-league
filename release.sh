#!/bin/sh

# restart docker container with new image version
cd $1;
RELEASE_TAG=$2 docker-compose up -d --build;

# prune all non-running images
docker image prune -af
