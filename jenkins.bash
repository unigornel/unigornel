#!/bin/bash

set -e

function error {
    echo error: "$@" >&2
    exit 1
}

[ -f "$GO_SSH_KEY" ] || error "file in GO_SSH_KEY not found: $GO_SSH_KEY"
[ -f "$MINIOS_SSH_KEY" ] || error "file in MINIOS_SSH_KEY not found: $MINIOS_SSH_KEY"

set -x
GIT_SSH="ssh -i $GO_SSH_KEY"        git submodule update --init go
GIT_SSH="ssh -i $MINIOS_SSH_KEY"    git submodule update --init minios
