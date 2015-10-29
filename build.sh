#!/bin/sh

export GOPATH="`pwd`/vendor:`pwd`"

go install bakapy/cmd/bakapy-scheduler && go install bakapy/cmd/bakapy-show-meta && go install bakapy/cmd/bakapy-run-job
