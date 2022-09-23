#!/bin/bash

set -e
set -x

go build fake.go
cp -fv fake ~/bin
