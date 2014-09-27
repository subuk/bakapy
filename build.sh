#!/bin/sh

export GOPATH="`pwd`/vendor:`pwd`"

go install bakapy-scheduler
