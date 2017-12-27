#!/bin/bash
set -e

docker run \
  --rm \
  --link postgres:postgres \
  --name migration $IMAGENAME \
  goose -path /db -env production up
