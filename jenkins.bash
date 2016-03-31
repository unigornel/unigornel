#!/bin/bash

set -e

function error {
    echo error: "$@" >&2
    exit 1
}

function do_cmd {
    echo "[+] $@"
    eval "$@"
}

function git_update_submodule_with_ssh {
    local name="$1"
    local file="$2"
    cat > ssh-wrapper.sh <<EOF
#!/bin/sh
ssh -i "$file" "\$@"
EOF
    chmod 0755 ssh-wrapper.sh
    do_cmd GIT_SSH=./ssh-wrapper.sh git submodule update --init "$name"
    rm ssh-wrapper.sh
}

[ -f "$GO_SSH_KEY" ] || error "file in GO_SSH_KEY not found: $GO_SSH_KEY"
[ -f "$MINIOS_SSH_KEY" ] || error "file in MINIOS_SSH_KEY not found: $MINIOS_SSH_KEY"

do_cmd git_update_submodule_with_ssh go     "$GO_SSH_KEY"
do_cmd git_update_submodule_with_ssh minios "$MINIOS_SSH_KEY"
