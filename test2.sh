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
$(pwd)/build/dmaster -dump -raw ${INF} > ${OUT}

if [[ x$? == x0 ]]; then
  echo result save to ${OUT}
fi
