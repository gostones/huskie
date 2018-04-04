#!/usr/bin/env bash

#args:
# CF_SPACE

set -x
#
export CF_SPACE=$1

#
base=$(cd "`dirname $0`"; pwd)
echo $base

keyfile="host_key"
rm -f $keyfile ${keyfile}.pub

##

[[ -z $HUSKIE_PORT ]] && export HUSKIE_PORT="2022"                    #chat server port
[[ -z $HUSKIE_IDENTITY ]] && export HUSKIE_IDENTITY="$base/$keyfile"  #chat server private key

if [ -z $CF_SPACE ]; then
    url="http://localhost:8022/tunnel"
else
    url="https://huskie-${CF_SPACE}.run.aws-usw02-pr.ice.predix.io/tunnel"
    [[ $http_proxy ]] && proxy="--proxy $http_proxy"
fi

ssh-keygen -f host_key -P ''

huskie client -v $proxy $url

##
