#!/bin/sh

export GOPATH="$GOPATH:`pwd`/vendor:`pwd`"
exec go test $@ bakapy
