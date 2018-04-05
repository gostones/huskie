#!/usr/bin/env bash

#args:
# cmd

set -x
#
usage() {
   cat << EOF

usage: $(basename $0) cmd
    dog
    mush
    whistle
EOF

exit 1
}

[[ $# -lt 1 ]] && usage

#
base=$(cd "`dirname $0`"; pwd)
echo $base

keyfile="host_key"
rm -f $keyfile ${keyfile}.pub

##

[[ -z $HUSKIE_PORT ]] && export HUSKIE_PORT="2022"                    #chat server port
[[ -z $HUSKIE_IDENTITY ]] && export HUSKIE_IDENTITY="$base/$keyfile"  #chat server private key

if [ $# -gt 1 ]; then
    url="http://localhost:8022/tunnel"
else
    url="https://huskie.run.aws-usw02-pr.ice.predix.io/tunnel"
    [[ $http_proxy ]] && proxy="--proxy $http_proxy"
fi

ssh-keygen -f host_key -P ''

huskie $1 -v $proxy $url

##
