#!/bin/bash

set -e

function do_cmd {
    echo "[+] $@" >&2
    eval -- "$@"
}

function die {
    echo "fatal:" "$@" >&2
    exit 1
}

function clean_submodule_commit {
    local path="$1"
    local contents=$(do_cmd git ls-tree HEAD "$path")
    local type=$(echo "$contents" | awk '{print $2}')
    local object=$(echo "$contents" | awk '{print $3}')

    [ "$type" = commit ] || die "object type is not a commit"
    echo "$object"
}

function dirty_submodule_commit {
    local path="$1"
    pushd "$path" >&2
    do_cmd git show-ref --head -s ^refs/origin/HEAD$
    popd >&2
}

function submodule_did_change {
    local clean=$(clean_submodule_commit "$1")
    local dirty=$(dirty_submodule_commit "$1")
    [ $clean != $dirty ] && return 0 || return 1
}

function changes_refs {
    do_cmd git rev-list ${1}..${2}
}

function show_ref {
    do_cmd git show --oneline "$1" | head -n 1
}

function submodule_changes {
    local submodule="$1"
    local clean=$(clean_submodule_commit "$submodule")
    local dirty=$(dirty_submodule_commit "$submodule")

    pushd "$submodule" >&2
    local refs=$(changes_refs $clean $dirty)
    for commit in $refs; do
        show_ref $commit
    done
    popd >&2
}

do_cmd git reset minios go

if ! submodule_did_change go && ! submodule_did_change minios; then
    die "No changes in submodules"
fi

COMMIT_MSG=/tmp/commit-$$
do_cmd touch $COMMIT_MSG
echo "Update submodules" >> $COMMIT_MSG
echo >> $COMMIT_MSG
if submodule_did_change go; then
    echo "Changes in go:" >> $COMMIT_MSG
    submodule_changes go >> $COMMIT_MSG
    echo >> $COMMIT_MSG
fi

if submodule_did_change minios; then
    echo "Changes in minios:" >> $COMMIT_MSG
    submodule_changes minios >> $COMMIT_MSG
    echo >> $COMMIT_MSG
fi

do_cmd git add go minios
do_cmd git commit -t $COMMIT_MSG
do_cmd rm $COMMIT_MSG
