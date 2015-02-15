#!/bin/bash

export PATH="$(dirname $(dirname `pwd`))/bin:$PATH"

taskId=8732d71b-077e-49ed-9222-b1177280de1e

bakapy-storage --config=bakapy.conf &
storagePid="$!"
bakapy-run-job --taskid="$taskId" --config=bakapy.conf --job=smoke

echo "TEST_OUTPUT: Waiting storage to exit"
kill $storagePid
while true; do
    test ! -d /proc/${storagePid} && echo "TEST_OUTPUT: Storage stopped" && exit 0
    sleep 1
done


if [ $(echo "$taskId"| wc -l) != 1 ];then
    echo "Error: bad taskId $taskId"
    exit 1
fi

mdSuccess=$(bakapy-show-meta --config=bakapy.conf --key=Success "$taskId")
if [ "$mdSuccess" != "true" ];then
    echo "Error: job failed"
    exit 1
fi

if [ ! -f storage/smoke/test_large.bin ];then
    echo "Error: file storage/smoke/test_large.bin not found"
    exit 1
fi

size=$(wc -c storage/smoke/test_large.bin |awk '{print $1}')
if [ "$size" != "10485760" ];then
    echo "Error: file storage/smoke/test_large.bin must be 10485760 bytes size, not $size"
    exit 1
fi

if [ ! -f storage/smoke/test1.txt ];then
    echo "Error: file storage/smoke/test1.txt not found"
    exit 1
fi

if [ $(cat storage/smoke/test1.txt) != "test1Content" ];then
    echo "Error: unexpected content in file storage/smoke/test1.txt"
    exit 1
fi
