#!/bin/sh

# change to project directory
cd $1;

# update shell script with current release data in case we need to manually restart the container later
echo "#!/bin/sh

# bring up any new containers
PL_APP_TAG=$2 docker compose up -d --build pl_app;

# restart existing containers
PL_APP_TAG=$2 docker compose restart pl_app;

if [ $? = 0 ]
then
    echo "version $2 released successfully!";
    exit 0;
else
    echo "failed to release $2";
    exit 1;
fi
" > re-release.sh

# make executable
chmod 0755 ./re-release.sh

# restart docker container with new image version
/bin/sh ./re-release.sh

# exit if failed
if [ $? != 0 ]; then
  exit 1
fi

# prune all non-running images
docker image prune -af
