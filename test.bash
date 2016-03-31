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
itime -v ./build.bash go
GOPATH=$PWD/minios/go itime -v ./build.bash app -a -x --app minios/go/src/sum/
