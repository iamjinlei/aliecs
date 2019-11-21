#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

OP=${1:-"desc"}
D=${2:-""}

go run $SCRIPT_DIR/../cmd/domain.go -op=$OP -domain=$D
