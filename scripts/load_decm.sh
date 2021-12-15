#!/bin/bash
rm -fv dmaster.tar.gz
scp hai.bai@10.9.112.11:/home/hai.bai/test/apps/decm/dmaster.tar.gz .
rm -rfv build
tar xvf dmaster.tar.gz
mv build/* .
rmdir build

