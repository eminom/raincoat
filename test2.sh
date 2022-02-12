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
time $(pwd)/build/dmaster -dump -raw -subr 7 ${INF} > ${OUT}

if [[ x$? == x0 ]]; then
  tail -10 ${OUT}
  echo result save to ${OUT}
fi

H1=$(sha256sum ${OUT}|awk '{print $1}')

OUT1="result1.txt"
time $(pwd)/build/dmaster -dump -raw -subr 1 ${INF} > ${OUT1}

H2=$(sha256sum ${OUT1}|awk '{print $1}')

if [[ x${H1} != x${H2} ]]; then
  echo ERROR NOT THE SAME
  echo H1 = ${H1}
  echo H2 = ${H2}
else
  echo SUCCESS the same hash ${H1}
fi

