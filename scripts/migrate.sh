#!/bin/bash
set -e

docker run \
  --rm \
  --network $NETWORK \
  --network-alias migration \
  --name migration $IMAGENAME \
  goose -path /db -env production up
