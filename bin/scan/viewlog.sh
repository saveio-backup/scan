#!/bin/bash
cd Log
logName=`ls -ltr | tail -1 | awk '{print $9}'`
cd ..
tail -f ./Log/${logName}
