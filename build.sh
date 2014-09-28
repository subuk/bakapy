#!/bin/sh

export GOPATH="`pwd`/vendor:`pwd`"

go install bakapy-scheduler && go install bakapy-show-meta
