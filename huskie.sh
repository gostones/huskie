#!/usr/bin/env bash

#args:
# cmd

set -x
#
usage() {
   cat << EOF

usage: $(basename $0) cmd
    pup
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

#export HUSKIE_URL="http://localhost:8080/tunnel"
export HUSKIE_URL="https://huskie.run.aws-usw02-pr.ice.predix.io/tunnel"

ssh-keygen -f host_key -P ''

huskie $@

##
