#!/bin/bash

make
#INF=$(readlink libprofile.data)
#if [[ -z $INF ]]; then
#  echo no such link 1>&2
#  exit 1
#fi

#echo decode for ${INF}
INF="raw_dpf.bin"
$(pwd)/build/dmaster -dump -raw ${INF}> result.txt
