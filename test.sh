#!/bin/bash

export GOPATH="$GOPATH:`pwd`/vendor:`pwd`"

for pkg in bakapy bakapy-storage; do
    go test -covermode=count -coverprofile="coverage-${pkg}.out" $@ "${pkg}"
done

echo 'mode: set' > coverage.out
for coverprof in $(ls -1 coverage-*.out);do
    cat $coverprof|grep -v 'mode:' >> coverage.out
done
