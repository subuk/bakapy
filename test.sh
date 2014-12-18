#!/bin/sh

export GOPATH="`pwd`/vendor:`pwd`"
exec go test bakapy
