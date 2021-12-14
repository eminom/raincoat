#!/bin/bash

make
time cat cluster_ringbuf_0.dump | awk '{print $2}' | $(pwd)/build/dmaster |sort -u
