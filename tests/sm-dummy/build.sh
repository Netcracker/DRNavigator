#!/bin/bash

cd "$(dirname "$0")" || exit 1

NAME=${NAME:-sm-dummy}

docker build -t "${NAME}.local" --no-cache .

for id in ${DOCKER_NAMES}
do
  docker tag ${NAME}.local ${id}
done
