#!/bin/bash

if [ "$BAKAPY_METADATA_SUCCESS" == "1" ];then
    exit 0
fi

if [ "$BAKAPY_EVENT" == "job_finished" ];then
sendmail "$BAKAPY_PARAM_MAILTO" <<EOF
To: root
Subject: [bakapy] Job $BAKAPY_METADATA_JOBNAME failed
Content-Type: text/plain; charset=utf8

$BAKAPY_METADATA_MESSAGE

Output:
-----------------------------
$BAKAPY_METADATA_OUTPUT
-----------------------------

Errput:
-----------------------------
$BAKAPY_METADATA_ERRPUT
-----------------------------
EOF
fi

if [ "$BAKAPY_EVENT" == "metadata_access_error" ];then
sendmail "$BAKAPY_PARAM_MAILTO" <<EOF
To: root
Subject: [bakapy] Job $BAKAPY_METADATA_JOBNAME failed
Content-Type: text/plain; charset=utf8

$BAKAPY_METADATA_MESSAGE

Output:
-----------------------------
$BAKAPY_METADATA_OUTPUT
-----------------------------

Errput:
-----------------------------
$BAKAPY_METADATA_ERRPUT
-----------------------------
EOF
fi
