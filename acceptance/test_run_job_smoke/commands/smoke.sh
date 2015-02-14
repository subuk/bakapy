#!/bin/bash

echo test1Content |_send_file  'test1.txt'
dd if=/dev/zero bs=1M count=10 |_send_file  'test_large.bin'
