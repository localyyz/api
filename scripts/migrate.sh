#!/bin/bash
set -e

docker run \
  --rm \
  --link postgres:postgres \
  --name migration $IMAGE\
  goose -env production up
