#!/bin/bash

gen_hostdata() {
cat <<EOF
package mimicdefs

// Update on $(date +%H%M%S_%Y%m%d)
// This file is generated automatically, do not change manually

$(go run mimic.go)

EOF
}

# Startup
FORCE=0
if [[ x$1 == xforce ]]; then
	FORCE=1
fi

OUTD="mimicdefs"

if [[ x${FORCE} == x1 ]]; then
	echo a force update
	rm -rfv ${OUTD}
fi


if [[ ! -d ${OUTD} ]]; then
	echo \# generating new definitions from proto types
	mkdir -p ${OUTD}
# one by one
	gen_hostdata > ${OUTD}/hostdata.go
	if [[ x$? != x0 ]]; then
		echo gen failed 1>&2
		exit 1
	fi
	ls -1 ${OUTD}
	gofmt -w ${OUTD}
	echo SUCCESS
fi


