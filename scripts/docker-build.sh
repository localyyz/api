#!/bin/bash
set -e

docker build -t paulx/api . && \
docker push paulx/api
