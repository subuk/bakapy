#!/bin/bash

set -e

DISTRO="$1"

ln -f -s "Dockerfile.${DISTRO}" Dockerfile
docker build -t "bakapy-build-${DISTRO}" .

rm -rf builded-packages/
mkdir -p builded-packages/
docker run --rm "bakapy-build-${DISTRO}" /bin/bash -c 'tar -C /packages -cf - .' | tar -C ./builded-packages/ -xf -
# docker stop -t 1 $CTID
# docker rm $CTID
