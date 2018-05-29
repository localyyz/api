#!/bin/bash
set -e

if [ ! -f $CONFIG ]; then
	echo "\"$CONFIG\" file missing"
	exit 1
fi

echo $SUP_NETWORK

docker run -d --name=$NAME \
  --network=$NETWORK \
  --network-alias=$NAME \
  -p 127.0.0.1:$HOST_PORT:$CONTAINER_PORT \
  -v $CONFIG:/etc/$NAME.conf \
  -v /data/etc/push.pem:/etc/push.pem \
  --restart=always \
  $IMAGENAME \
  /bin/$NAME -config=/etc/$NAME.conf
