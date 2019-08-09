#!/bin/bash
logName=`ls -ltr Log | tail -1 | awk '{print $9}'`
if [[ $1 == "-o" ]]; then
	echo $logName
	exit 0
fi

tail -f ./Log/${logName}
