#!/bin/bash
set -e

if [ $(docker-machine status) != "Running" ]; then
  docker-machine start
else
  echo docker-machine running.
fi

eval $(docker-machine env default)
docker build -t paulx/api . && \
docker push paulx/api
