#!/usr/bin/env bash

set -x

base=$(cd "`dirname $0`"; pwd)
echo $base
#

[[ -z $PORT ]] && export PORT="8080"

export PATH=/app/bin:$PATH

pwd
env

#
huskie harness

echo "Server stopped"
##