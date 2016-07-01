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
INTEGRATION=y
FAST=n
SHOW_HELP=n
SETUP_GOPATH=n

TEMP=`getopt -o h --long help,no-go,no-integration,fast,setup-gopath -n test.bash -- "$@"`
eval set -- "$TEMP"
while true; do
    case "$1" in
        -h|--help)      SHOW_HELP=y         ; shift   ;;
        --no-go)        BUILD_GO=n          ; shift   ;;
        --no-integration) INTEGRATION=n     ; shift   ;;
        --fast)         FAST=y              ; shift   ;;
        --setup-gopath) SETUP_GOPATH=y      ; shift   ;;
        --)             shift               ; break   ;;
    esac
done

if [ "$SHOW_HELP" = y ]; then
    echo "usage: test.bash -h|--help"
    echo "       test.bash [--no-go] [--no-integration] [--fast] [--setup-gopath]"
    exit 0
fi

function do_cmd {
    echo "[+] $@"
    eval "$@"
}

PROJECT_ROOT="$PWD"
if [ "$SETUP_GOPATH" = y ]; then
    do_cmd rm -rf gopath
    do_cmd mkdir -p gopath/src/github.com/unigornel
    do_cmd ln -s ../../../.. gopath/src/github.com/unigornel/unigornel
    export GOPATH="$PWD/gopath"
    do_cmd cd gopath/src/github.com/unigornel/unigornel
fi

# Build Go
if [ "$BUILD_GO" = y ]; then
    pushd go/src
    export GOROOT_BOOTSTRAP
    [ -z "$GOROOT_BOOTSTRAP" ] && error "GOROOT_BOOTSTRAP not set"
    [ "$FAST" = y ] && fast_opt=--no-clean || fast_opt=
    do_cmd GOOS=unigornel GOARCH=amd64 ./make.bash $fast_opt
    popd
fi

# Build unigornel
pushd unigornel
do_cmd go get -v
do_cmd go build -o unigornel
cat > .unigornel.yaml <<EOF
goroot: $PWD/../go
minios: $PWD/../minios
EOF
eval $(./unigornel env -c .unigornel.yaml)
export PATH="$PWD:$PATH"
popd

# Run integration tests
if [ "$INTEGRATION" = y ]; then
    unigornel_root="$PWD"
    pushd integration_tests
    do_cmd go get -v
    go build -o "test"
    ./test -junit "integration_tests.xml"
    popd
fi
