#!/usr/bin/env python3

import sys

def xMain(infile):
    with open(infile) as fin:
        for line in fin.readlines():
            line = line.strip()
            if line.startswith("TS") or \
             line.startswith("CQM") and \
             (line.find("event=9") >= 0 or line.find("event=8") >= 0):
                print(line)

if '__main__' == __name__:
    if len(sys.argv) < 2:
        print("need input")
        sys.exit(1)
    xMain(sys.argv[1])
