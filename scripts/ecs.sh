#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

OP=${1:-"desc"}
IDX=${2:-0}
N=$((IDX+1))

if [ $OP = "up" ] || [ $OP = "down" ] || [ $OP = "del" ] || [ $OP = "desc" ] || [ $OP = "run" ] || [ $OP = "proxy" ]; then
	go run $SCRIPT_DIR/../cmd/ecs.go -op=$OP -idx=$IDX
	if [ $OP = "proxy" ]; then	
		ip=$(go run $SCRIPT_DIR/../cmd/ecs.go -op=desc | grep -v "\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-" | grep -v "Public IP" | head -n $IDX | tail -n 1 | cut -d"|" -f8 | xargs)
		expect -c 'spawn ssh -o StrictHostKeyChecking=no -CN -D 8080 root@'"$ip"'; expect "assword:"; send "'"$ECS_ROOT_PWD"'\r"; interact'
	fi
elif [ $OP = "go" ]; then
	ip=$(go run $SCRIPT_DIR/../cmd/ecs.go -op=desc | grep -v "\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-" | grep -v "Public IP" | head -n $N | tail -n 1 | cut -d"|" -f8 | xargs)
    expect -c 'spawn ssh -o StrictHostKeyChecking=no root@'"$ip"'; expect "assword:"; send "'"$ECS_ROOT_PWD"'\r"; interact'
else
	echo -e "supported commands are: up, down, del, desc\n"
fi
