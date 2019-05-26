#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if [ $1 = "up" ] || [ $1 = "down" ] || [ $1 = "del" ] || [ $1 = "desc" ]; then
	go run $SCRIPT_DIR/../cmd/ecs.go -op=$1
else
	echo -e "supported commands are: up, down, del, desc\n"
fi
