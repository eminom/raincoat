#!/bin/bash

pushd .
TARGETD=$(dirname $0)
cd ${TARGETD}
echo now update in $(pwd)
if [[ x$? != x0 ]]; then
  echo switch error 1>&2
  exit 1
fi

rm -fv dmaster.tar.gz
scp hai.bai@10.9.112.11:/home/hai.bai/test/apps/decm/dmaster.tar.gz .
rm -rfv build
tar xvf dmaster.tar.gz
mv build/* .
rmdir build


popd
