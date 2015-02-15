#!/bin/bash

for runTest in $(find acceptance/ -name 'run.sh');do
    pushd $(dirname $runTest) >/dev/null
    test -x clean.sh && ./clean.sh

    if [ "$1" == "-v" ];then
        testOutput=""
        ./run.sh
        testRet=$?
    else
        testOutput=$(bash run.sh 2>&1)
        testRet=$?
    fi

    test -x clean.sh && ./clean.sh
    if [ "$testRet" != 0 ];then
        echo "fail	$(dirname $runTest). See output below."
        echo -e "$testOutput"
        exit 1
    fi
    popd >/dev/null
    echo "ok	$(dirname $runTest)"
done
