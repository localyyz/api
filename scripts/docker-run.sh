#!/bin/bash
set -e

if [ ! -f $CONFIG ]; then
	echo "\"$CONFIG\" file missing"
	exit 1
fi

docker run -d \
  -p $HOST_PORT:$CONTAINER_PORT \
  -v $CONFIG:/etc/api.conf \
  -v $BINARY:/etc/Moodie.ipa \
  --link postgres:postgres \
  --restart=always \
  --name $NAME $IMAGE
