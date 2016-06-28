#!/bin/bash

set -e

PR_REFSPEC='+refs/pull/*:refs/remotes/origin/pr/*'

add_pr_refs() {
    if ! git config --get-all remote.origin.fetch | grep -F "$PR_REFSPEC" > /dev/null; then
        git config --add remote.origin.fetch "$PR_REFSPEC"
    fi
}

pushd go && add_pr_refs && popd
pushd minios && add_pr_refs && popd

git submodule update --init --recursive

opts=
if [ "$1" = "--fast" ]; then
    opts=--fast
fi

taskset -c "${EXECUTOR_NUMBER}" ./test.bash --setup-gopath $opts
