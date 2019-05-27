#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if [ $1 = "up" ] || [ $1 = "down" ] || [ $1 = "del" ] || [ $1 = "desc" ]; then
	go run $SCRIPT_DIR/../cmd/ecs.go -op=$1
elif [ $1 = "go" ]; then
	ip=$(go run $SCRIPT_DIR/../cmd/ecs.go -op=desc | grep Running | cut -d"|" -f5 | xargs)
    expect -c 'spawn ssh root@'"$ip"'; expect "assword:"; send "'"$ECS_ROOT_PWD"'\r"; interact'
else
	echo -e "supported commands are: up, down, del, desc\n"
fi
