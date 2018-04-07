#!/usr/bin/env bash

##
[[ $DEBUG ]] && FLAG="-x"

function build() {
    mkdir -p bin
    rm -rf bin/*

    echo "## Cleaning ..."
    go clean $FLAG ./...

    echo "## Vetting ..."
    go vet $FLAG ./...; if [ $? -ne 0 ]; then
        return 1
    fi

    echo "## Building ..."
#    go build $FLAG -buildmode=exe -o bin/huskie -ldflags '-extldflags "-static"'; if [ $? -ne 0 ]; then
#        return 1
#    fi
    go build $FLAG -buildmode=exe -o bin/huskie; if [ $? -ne 0 ]; then
        return 1
    fi

    echo "## Testing ..."
    go test $FLAG ./...; if [ $? -ne 0 ]; then
        return 1
    fi
}

echo "#### Building ..."

build; if [ $? -ne 0 ]; then
    echo "#### Build failure"
    exit 1
fi

echo "#### Build success"

exit 0