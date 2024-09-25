#!/bin/bash -e

GOTRACEBACK=crash GODEBUG=gctrace=1 ./gogc1.16.5 --ballast=true
#GOTRACEBACK=crash ./gogc1.16.5 --ballast=false
