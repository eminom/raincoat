#!/bin/bash

make

if [[ x$? != x0 ]]; then
  echo build failed 1>&2
  exit 1
fi
#INF=$(readlink libprofile.data)
#if [[ -z $INF ]]; then
#  echo no such link 1>&2
#  exit 1
#fi

#echo decode for ${INF}
INF="raw_dpf.bin"
OUT="process.txt"
$(pwd)/build/dmaster -proc ${INF} -meta /home/hai.bai/data16/meta | tee ${OUT}

if [[ x$? == x0 ]]; then
  echo result save to ${OUT}
else
  echo test failed in $? 1>&2
fi
