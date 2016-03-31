#!/bin/bash

set -e

HELP=n
DID_ANYTHING=n
GOROOT_BOOTSTRAP=
MINIOS_GOARCHIVE=
MINIOS_GOINCLUDE=

TEMP=`getopt -o h --long help,gobootstrap:,goarchive:,goinclude: -n build.bash -- "$@"`
eval set -- "$TEMP"
while true; do
    case "$1" in
        -h|--help)
            HELP=y
            shift
            ;;
        --gobootstrap)
            GOROOT_BOOTSTRAP="$2"
            shift 2
            ;;
        --goarchive)
            MINIOS_GOARCHIVE="$2"
            shift 2
            ;;
        --goinclude)
            MINIOS_GOINCLUDE="$2"
            shift 2
            ;;
        --)
            shift
            break
            ;;
    esac
done

function do_cmd {
    echo "[+] $@"
    eval "$@"
}

# Show help
function usage {
    echo "usage: build.bash -h|--help"
    echo "       build.bash [--gobootstrap path]"
    echo "                  [--goarchive app/app.a --goinclude app/]"
}

if [ "$HELP" = y ]; then
    usage
    exit 0
fi

# Build Go
if [ -n "$GOROOT_BOOTSTRAP" ]; then
    DID_ANYTHING=y

    pushd go/src
    echo "Building Go in $PWD"
    do_cmd GOROOT_BOOTSTRAP="$GOROOT_BOOTSTRAP" ./make.bash
    popd
fi

# Build Mini-OS
if [ -n "$MINIOS_GOARCHIVE" -a -n "$MINIOS_GOINCLUDE" ]; then
    DID_ANYTHING=y

    archive="$(realpath "$MINIOS_GOARCHIVE")"
    include="$(realpath "$MINIOS_GOINCLUDE")"

    pushd minios
    echo "Building Mini-OS in $PWD"
    do_cmd make GOARCHIVE="$archive" GOINCLUDE="$include"
    popd

elif [ -n "$MINIOS_GOARCHIVE" -o -n "$MINIOS_GOINCLUDE" ]; then
    usage >&2
    exit 1
fi

# Make sure something is done
if [ "$DID_ANYTHING" = n ]; then
    usage >&2
    exit 1
fi
