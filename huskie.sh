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

##

export HUSKIE_URL="https://huskie.run.aws-usw02-pr.ice.predix.io/tunnel"

huskie $@

##
