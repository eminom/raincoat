#!/usr/bin/env python3

import sys


def xMain():
  with open('syncpoints.txt') as fin:
    syncVec = []
    hostMin, hostMax = 1<<64, 0
    cycleMin, cycleMax = 1<<64, 0
    for line in fin.readlines():
      vs = line.split()
      if len(vs) != 3:
        continue
      (dpfSyncIndex, hostTime, devCycle) = map(int, vs)
      syncVec.append((dpfSyncIndex, hostTime, devCycle))
      hostMax = max(hostMax, hostTime)
      cycleMax = max(cycleMax, devCycle)
      hostMin = min(hostMin, hostTime)
      cycleMin = min(cycleMin, devCycle)
    
    print("host min: ", hostMin)
    print("host max: ", hostMax)
    print("dev min: ", cycleMin)
    print("dev max: ", cycleMax)
    print("host span: ", hostMax - hostMin)
    print("cycle span: ", cycleMax - cycleMin)
    print("host span in seconds: ", (hostMax - hostMin) / (1000*1000*1000))

if '__main__' == __name__:
  xMain()
