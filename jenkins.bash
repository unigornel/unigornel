#!/bin/bash

set -e

opts=
if [ "$1" = "--fast" ]; then
    opts=--fast
fi

taskset -c "${EXECUTOR_NUMBER}" ./test.bash --setup-gopath $opts
