#!/bin/bash

set -e

DISTRO="$1"

if [ -z "$DISTRO" ];then
    echo "Please specify distro as first argument. Choices are:"
    for f in $(ls -1 Dockerfile.*);do
        echo '-' ${f:11}
    done
    exit 1
fi

rm -f Dockerfile
ln -s "Dockerfile.${DISTRO}" Dockerfile
docker build -t "bakapy-build-${DISTRO}" .

rm -rf "native-packages/${DISTRO}"
mkdir -p "native-packages/${DISTRO}"
docker run --rm "bakapy-build-${DISTRO}" /bin/bash -c 'tar -C /packages -cf - .' | tar -C "./native-packages/${DISTRO}" -xf -
