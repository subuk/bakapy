#!/bin/bash

export PATH="$(dirname $(dirname `pwd`))/bin:$PATH"
trap 'trap - SIGTERM && kill -INT 0' SIGINT SIGTERM EXIT

taskId=8732d71b-077e-49ed-9222-b1177280de1e

bakapy-scheduler &>scheduler.log &
schedulerPid="$!"

echo -n "Waiting scheduler to start ..."
i=0
while true; do
    let i+=1
    netstat -tlnp 2>/dev/null |awk '{print $4}' |grep -q :19875$ && break
    echo -n "."
    if [ "$i" -gt 50 ];then
        echo "Error: timeout"
        exit 1
    fi
    sleep 0.1
done
echo " OK"

bakapy-storage &>storage.log &
storagePid="$!"

echo -n "Waiting storage to start ..."
i=0
while true; do
    let i+=1
    netstat -tlnp 2>/dev/null |awk '{print $4}' |grep -q :19876$ && break
    echo -n "."
    if [ "$i" -gt 50 ];then
        echo "Error: timeout"
        exit 1
    fi
    sleep 0.1
done
echo " OK"


echo -n 'Waiting bakapy-run-job to finish ... '
bakapy-run-job --taskid="$taskId" --config=bakapy.conf --job=smoke &>run-job.log
runJobExitCode="$?"
echo "OK: retcode $runJobExitCode"

kill $storagePid
echo -n "Waiting storage to exit ... "
while true; do
    test ! -d /proc/${storagePid} && break
    echo -n ". "
    sleep 0.1
done
echo "OK"

kill $schedulerPid
echo -n "Waiting scheduler to exit ... "
while true; do
    test ! -d /proc/${schedulerPid} && break
    echo -n ". "
    sleep 0.1
done
echo "OK"

echo -n "Checking bakapy-run-job exit code ... "
if [ "$runJobExitCode" != "0" ];then
    echo "Error: bakapy-run-job exited with code $runJobExitCode"
    echo "------- RUN JOB OUTPUT --------"
    cat run-job.log
    echo "----- END RUN JOB OUTPUT ------"
    echo "-------  SCHED OUTPUT  --------"
    cat scheduler.log
    echo "-----  END SCHED OUTPUT  ------"
    exit 1
fi
echo OK

echo -n "Checking job finished successfully ... "
mdSuccess=$(bakapy-show-meta --config=bakapy.conf --key=Success "$taskId")
if [ "$mdSuccess" != "true" ];then
    echo "Failed: $(bakapy-show-meta --config=bakapy.conf --key=Message "$taskId")"
    exit 1
fi
echo OK

echo -n "Checking file storage/smoke/test_large.bin exist ... "
if [ ! -f storage/smoke/test_large.bin ];then
    echo "Error: file storage/smoke/test_large.bin not found"
    exit 1
fi
echo OK

echo -n "Checking file storage/smoke/test_large.bin size ... "
size=$(wc -c storage/smoke/test_large.bin |awk '{print $1}')
if [ "$size" != "5242880" ];then
    echo "Error: file storage/smoke/test_large.bin must be 5242880 bytes size, not $size"
    exit 1
fi
echo "OK: $size bytes"

echo -n "Checking file storage/smoke/test1.txt exist ... "
if [ ! -f storage/smoke/test1.txt ];then
    echo "Error: file storage/smoke/test1.txt not found"
    exit 1
fi
echo OK

echo -n "Checking file storage/smoke/test1.txt content ... "
if [ $(cat storage/smoke/test1.txt) != "test1Content" ];then
    echo "Error: unexpected content in file storage/smoke/test1.txt"
    exit 1
fi
echo OK

echo -n "Checking notification script runned ... "
if [ "$(cat NOTIFICATION)" != "Job smoke finished" ];then
    echo "Error: no notification send"
    exit 1
fi
echo OK
