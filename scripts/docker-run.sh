#!/bin/bash
set -e

if [ ! -f $CONFIG ]; then
	echo "\"$CONFIG\" file missing"
	exit 1
fi

docker run -d --name=$NAME \
  -p 127.0.0.1:$HOST_PORT:$CONTAINER_PORT \
  -v $CONFIG:/etc/$NAME.conf \
  -v /data/etc/push.pem:/etc/push.pem \
  --link postgres:postgres \
  --restart=always \
  $IMAGE \
  /bin/$NAME -config=/etc/$NAME.conf
