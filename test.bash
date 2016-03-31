#!/bin/bash

function error {
    echo error: "$@"
    exit 1
}

[ -z "$GOROOT_BOOTSTRAP" ] && error "GOROOT_BOOTSTRAP not set"
export GOROOT_BOOTSTRAP

./build.bash go
./build.bash app -a -x --app minios/go/src/sum/
