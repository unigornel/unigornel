#!/bin/bash

set -e

function error {
    echo error: "$@"
    exit 1
}

function itime {
    /usr/bin/time "$@"
}

[ -z "$GOROOT_BOOTSTRAP" ] && error "GOROOT_BOOTSTRAP not set"
export GOROOT_BOOTSTRAP

set -x
#itime -p ./build.bash go
GOPATH=$PWD/minios/go itime -p ./build.bash app -a -x --app minios/go/src/sum/
