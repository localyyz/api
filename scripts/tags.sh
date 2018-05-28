#!/bin/bash -e

function usage() {
  echo "Usage: $0 <operation>"
  echo "-"
  echo "operation : [ list, clean ]"
  exit 1
}

function list() {
  echo "LISTING IMAGE TAGS";
  gcloud container images list-tags $IMAGE
}


function clean() {
  echo "CLEAN OLD IMAGE TAGS";
  gcloud container images list-tags $IMAGE \
    | grep $ENV_BRANCH \
    | tail -n +4 \
    | awk '{print $1}' \
    | xargs -I imagetag bash -c 'gcloud container images delete $IMAGE@sha256:imagetag --force-delete-tags';

#gcloud container images delete gcr.io/<project>/api@sha256:<tag> --force-delete-tags
}

if [ $# -lt 1 ]; then
  usage
fi

operation="$1"

case "$operation" in
  "list")
    list
    ;;
  "clean")
    clean
    ;;
  *)
    echo "no such operation"
    usage
    ;;
esac

