#!/bin/bash

set -e

. ../tools.sh
export PATH="$(dirname $(dirname `pwd`))/bin:$PATH"
trap 'trap - SIGTERM && kill -INT 0' SIGINT SIGTERM EXIT

taskId=8732d71b-077e-49ed-9222-b1177280de1e

bakapy-metaman &>metaman.log &
metamanPid="$!"

echo -n "Waiting metaman to start ..."
wait_port_listened 19875
echo " OK"

on_exit(){
    echo -n "Waiting metaman to exit ..."
    kill_and_wait "$metamanPid"
    echo " OK"
}

trap on_exit EXIT

echo -n 'Running storage cleanup ...'
bakapy-storage --clean-only &>storage.log
if [ "$?" == "0" ];then
    echo " OK"
else
    echo " Fail"
    cat storage.log >&2
    exit 1
fi

echo -n "Checking expired job metadata does not exist ... "
if [ -f "metadata/73e00f56-48b2-4347-8026-390239863ab2" ];then
    echo "Failed: metadata/73e00f56-48b2-4347-8026-390239863ab2  present"
    exit 1
fi
echo "OK"

echo -n "Checking files for expired metadata does not exist ... "
if [ -f "storage/ns1/expired-test10.txt" ];then
    echo "Failed: storage/ns1/expired-test10.txt exists"
    exit 1
fi
echo "OK"

echo -n "Checking invalid metadata still exists ... "
if [ ! -f "metadata/e848a587-0a06-4269-b0be-3e5d0bab2f7d" ];then
    echo "Failed: metadata/e848a587-0a06-4269-b0be-3e5d0bab2f7d removed"
    exit 1
fi
echo "OK"

echo -n "Checking not expired metadata still exists ... "
if [ ! -f "metadata/8a07f445-afde-41ee-8912-0788c3f02699" ];then
    echo "Failed: metadata/8a07f445-afde-41ee-8912-0788c3f02699 removed"
    exit 1
fi
echo "OK"

echo -n "Checking not expired files exists ... "
if [ ! -f "storage/ns1/test1.txt" ];then
    echo "Failed: storage/ns1/test1.txt"
    exit 1
fi
if [ ! -f "storage/ns1/test2.txt" ];then
    echo "Failed: storage/ns1/test1.txt"
    exit 1
fi
echo "OK"
