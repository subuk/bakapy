#!/bin/bash

echo test1Content |_send_file  'test1.txt'
dd if=/dev/urandom bs=1M count=5 |_send_file  'test_large.bin'
