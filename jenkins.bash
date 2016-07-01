#!/bin/bash

set -e

PR_REFSPEC='+refs/pull/*:refs/remotes/origin/pr/*'

err() {
    echo "$@" >&2
    exit 1
}

add_pr_refs() {
    if ! git config --get-all remote.origin.fetch | grep -F "$PR_REFSPEC" > /dev/null; then
        git config --add remote.origin.fetch "$PR_REFSPEC"
    fi
}

[ -n "$EXECUTOR_NUMBER" ] || err "error: EXECUTOR_NUMBER not set"

pushd go && add_pr_refs && popd
pushd minios && add_pr_refs && popd

git submodule update --init --recursive

opts=
if [ "$1" = "--fast" ]; then
    opts=--fast
fi


export TEST_PING_NETWORK="10.123.${EXECUTOR_NUMBER}.0/24"
taskset -c "${EXECUTOR_NUMBER}" ./test.bash --setup-gopath $opts
