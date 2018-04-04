#!/usr/bin/env bash

#args:
# CF_SPACE

#DRY_RUN=echo

#
usage() {
   cat << EOF

usage: $(basename $0) CF_SPACE

EOF

exit 1
}

[[ $# -lt 1 ]] && usage

#
export CF_SPACE=$1

##
function deploy() {
    local domain="run.aws-usw02-pr.ice.predix.io"
    local name="huskie-${CF_SPACE}"

    echo "Pushing service $domain $name ..."

    $DRY_RUN cf delete-route $domain --hostname $name -f

    $DRY_RUN cf push -f manifest.yml -d $domain --hostname $name; if [ $? -ne 0 ]; then
        return 1
    else
        return 0
    fi
}

#
echo "### Deploying $CF_SPACE ..."

deploy; if [ $? -ne 0 ]; then
    echo "#### Deploy failed"
    exit 1
fi

exit 0
##
