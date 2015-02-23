#!/bin/bash

set -e

DISTRO="$1"
SIGN_KEY="$2"
USAGE="$0 <distro> <key>"


if [ -z "$DISTRO" ];then
    echo "Please specify distro as first argument. Choices are:"
    for f in $(ls -1 Dockerfile.*);do
        echo '-' ${f:11}
    done
    exit 1
fi

if [ -z "$SIGN_KEY" ];then
    echo "Please specify signing gpg key as second argument"
    exit 1
fi

docker build --rm -t "bakapy-build-${DISTRO}" -f Dockerfile.${DISTRO} .

rm -rf "native-packages/${DISTRO}"
mkdir -p "native-packages/${DISTRO}"
docker run --rm "bakapy-build-${DISTRO}" /bin/bash -c 'tar -C /packages -cf - .' | tar -C "./native-packages/${DISTRO}" -xf -

for distro in precise trusty wheezy;do
    if [ "$DISTRO" != $distro ];then
        continue
    fi
    cd "native-packages/${DISTRO}"
    dpkg-sig -k "$SIGN_KEY" --sign builder --sign-changes full *.changes
done

for distro in centos6;do
    if [ "$DISTRO" != $distro ];then
        continue
    fi
    cd "native-packages/${DISTRO}"
    rpm --resign *.rpm
done
