#!/bin/bash

set -e

taskset -c "${EXECUTOR_NUMBER}" ./test.bash --setup-gopath
