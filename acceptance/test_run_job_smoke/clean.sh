#!/bin/bash

rm -f storage/smoke/test_large.bin
rm -f storage/smoke/test1.txt
test -d storage/smoke && rmdir storage/smoke
rm -f metadata/8732d71b-077e-49ed-9222-b1177280de1e
rm -f NOTIFICATION
rm -f run-job.log
rm -f scheduler.log
rm -f storage.log
rm -f metaman.log
