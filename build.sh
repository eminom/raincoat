#!/bin/bash
set +e
go build dmaster.go

REMOTE=hai.bai@10.12.101.31
REMOTED=/home/hai.bai/bin

ssh ${REMOTE} rm -rfv /home/hai.bai/bin/dmaster
scp dmaster ${REMOTE}:${REMOTED}

ssh ${REMOTE} rm -rfv /home/hai.bai/bin/xdmaster
scp xdmaster ${REMOTE}:${REMOTED}

echo update successfully
