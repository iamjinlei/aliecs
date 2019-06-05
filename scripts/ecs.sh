#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

OP=${1:-"desc"}

if [ $OP = "up" ] || [ $OP = "down" ] || [ $OP = "del" ] || [ $OP = "desc" ] || [ $OP = "run" ] || [ $OP = "proxy" ]; then
	go run $SCRIPT_DIR/../cmd/ecs.go -op=$OP
elif [ $OP = "go" ]; then
	IDX=$((${2:-0}+1))
	echo "IDX = $IDX"
	ip=$(go run $SCRIPT_DIR/../cmd/ecs.go -op=desc | grep -v "\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-" | grep -v "Public IP" | head -n $IDX | tail -n 1 | cut -d"|" -f7 | xargs)
    expect -c 'spawn ssh -o StrictHostKeyChecking=no root@'"$ip"'; expect "assword:"; send "'"$ECS_ROOT_PWD"'\r"; interact'
else
	echo -e "supported commands are: up, down, del, desc\n"
fi
