#!/bin/bash

make
#INF=$(readlink libprofile.data)
#if [[ -z $INF ]]; then
#  echo no such link 1>&2
#  exit 1
#fi

#echo decode for ${INF}
INF="raw_dpf.bin"
OUT="result.txt"
rm -rfv ${OUT}
time $(pwd)/build/dmaster -dump -raw ${INF} > ${OUT}

if [[ x$? == x0 ]]; then
  tail -10 ${OUT}
  echo result save to ${OUT}
fi
