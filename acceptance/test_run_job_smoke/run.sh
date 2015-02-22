#!/bin/bash

export PATH="$(dirname $(dirname `pwd`))/bin:$PATH"
trap 'trap - SIGTERM && kill -INT 0' SIGINT SIGTERM EXIT

taskId=8732d71b-077e-49ed-9222-b1177280de1e

bakapy-metaman &>metaman.log &
metamanPid="$!"

echo -n "Waiting metaman to start ..."
wait_port_listened 19875
echo " OK"

bakapy-storage &>storage.log &
storagePid="$!"

echo -n "Waiting storage to start ..."
wait_port_listened 19876
echo " OK"

on_exit(){
    echo -n "Waiting storage to exit ... "
    kill_and_wait "$storagePid"
    echo "OK"

    echo -n "Waiting metaman to exit ... "
    kill_and_wait "$metamanPid"
    echo "OK"
}

trap on_exit EXIT

echo -n 'Waiting bakapy-run-job to finish ... '
bakapy-run-job --taskid="$taskId" --job=smoke &>run-job.log
runJobExitCode="$?"
echo "OK: retcode $runJobExitCode"

echo -n "Checking bakapy-run-job exit code ... "
if [ "$runJobExitCode" != "0" ];then
    echo "Error: bakapy-run-job exited with code $runJobExitCode"
    echo "------- RUN JOB OUTPUT --------"
    cat run-job.log
    echo "----- END RUN JOB OUTPUT ------"
    echo "-------  SCHED OUTPUT  --------"
    cat metaman.log
    echo "-----  END SCHED OUTPUT  ------"
    exit 1
fi
echo OK

echo -n "Checking job finished successfully ... "
mdSuccess=$(bakapy-show-meta --key=Success "$taskId")
if [ "$mdSuccess" != "true" ];then
    echo "Failed: $(bakapy-show-meta --key=Message "$taskId")"
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
