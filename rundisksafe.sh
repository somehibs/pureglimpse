#!/bin/bash

# build and run golang app called pure
go build -o pure cli/main.go
./pure &
while true; do
	# gets the size of current dir
	left=`df -m . | awk '{print $4}' | grep -v Available`
	if [[ $left -lt 10000 ]]; then
		echo "Less than 10gb remaining, killing"
		killall pure
	else
		echo "$left remaining"
	fi
	sleep 5
done
