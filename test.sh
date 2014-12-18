#!/bin/sh

export GOPATH="`pwd`/vendor:`pwd`"

go test bakapy
