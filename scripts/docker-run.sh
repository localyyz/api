#!/bin/bash
set -e

if [ ! -f $CONFIG ]; then
	echo "\"$CONFIG\" file missing"
	exit 1
fi

docker run -d \
  -p $HOST_PORT:$CONTAINER_PORT \
  -v $CONFIG:/etc/api.conf \
  -v /data/etc/push.pem:/etc/push.pem \
  --link postgres:postgres \
  --restart=always \
  --name $NAME $IMAGE
