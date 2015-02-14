#!/bin/bash

for runTest in $(find acceptance/ -name 'run.sh');do
    pushd $(dirname $runTest)
    test -x clean.sh && ./clean.sh
    bash run.sh
    testRet=$?
    test -x clean.sh && ./clean.sh
    if [ "$testRet" != 0 ];then
        echo "Failed $runTest"
        exit 1
    fi
    popd >/dev/null
done
