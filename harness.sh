#!/usr/bin/env bash

set -x

#base=$(cd "`dirname $0`"; pwd)
#echo $base
#
#keyfile="host_key"
#rm -f $keyfile ${keyfile}.pub

##

#[[ -z $HUSKIE_PORT ]] && export HUSKIE_PORT="2022"                    #chat server port
#[[ -z $HUSKIE_IDENTITY ]] && export HUSKIE_IDENTITY="$base/$keyfile"  #chat server private key

[[ -z $PORT ]] && export PORT="8080"

export PATH=/app/bin:$PATH

pwd
env

##
#ssh-keygen -f host_key -P ''

huskie harness

echo "Server stopped"
##