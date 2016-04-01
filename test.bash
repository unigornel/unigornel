#!/bin/bash

set -e

function error {
    echo error: "$@"
    exit 1
}

function itime {
    /usr/bin/time "$@"
}

BUILD_GO=y
SHOW_HELP=n

TEMP=`getopt -o h --long help,no-go -n test.bash -- "$@"`
eval set -- "$TEMP"
while true; do
    case "$1" in
        -h|--help)      SHOW_HELP=y         ; shift   ;;
        --no-go)        BUILD_GO=n          ; shift   ;;
        --)             shift               ; break   ;;
    esac
done

if [ "$SHOW_HELP" = y ]; then
    echo "usage: test.bash -h|--help"
    echo "       test.bash [--no-go]"
    exit 0
fi

function do_cmd {
    echo "[+] $@"
    eval "$@"
}

# Build Go
if [ "$BUILD_GO" = y ]; then
    export GOROOT_BOOTSTRAP
    [ -z "$GOROOT_BOOTSTRAP" ] && error "GOROOT_BOOTSTRAP not set"
    do_cmd itime -p ./build.bash go
fi

# Build test application
GOPATH=$PWD/integration_tests/hello_world/go do_cmd itime -p ./build.bash app -a -x --app integration_tests/hello_world/go/src/helloworld/
