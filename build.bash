#!/bin/bash

set -e

HELP=n
DID_ANYTHING=n
[ -z "$GOROOT_BOOTSTRAP" ] && GOROOT_BOOTSTRAP=
MINIOS_GOARCHIVE=
MINIOS_GOINCLUDE=
APP_DIR=
O_FLAG=
A_FLAG=n
X_FLAG=n

TEMP=`getopt -o ho:ax --long help,gobootstrap:,goarchive:,goinclude:,app: -n build.bash -- "$@"`
eval set -- "$TEMP"
while true; do
    case "$1" in
        -h|--help)      HELP=y                  ; shift   ;;
        --gobootstrap)  GOROOT_BOOTSTRAP="$2"   ; shift 2 ;;
        --goarchive)    MINIOS_GOARCHIVE="$2"   ; shift 2 ;;
        --goinclude)    MINIOS_GOINCLUDE="$2"   ; shift 2 ;;
        --app)          APP_DIR="$2"            ; shift 2 ;;
        -o)             O_FLAG="$2"             ; shift 2 ;;
        -a)             A_FLAG=y                ; shift   ;;
        -x)             X_FLAG=y                ; shift   ;;
        --)             shift                   ; break   ;;
    esac
done

function do_cmd {
    echo "[+] $@"
    eval "$@"
}

function error {
    echo error: "$@" >&2
    exit 1
}

# Show help
function usage {
    echo "usage: build.bash -h|--help"
    echo "       build.bash --gobootstrap path"
    echo "       build.bash --goarchive app/app.a --goinclude app/"
    echo "       build.bash --app path/to/app/dir/ [-o path/to/app.a]"
    echo "                  [-a] [-x]"
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
    do_cmd GOROOT_BOOTSTRAP="$GOROOT_BOOTSTRAP" GOOS=netbsd GOARCH=amd64 ./make.bash
    popd
fi

# Build application
if [ -n "$APP_DIR" ]; then
    [ -z "$GOPATH" ] && error "cannot build app: GOPATH not set"

    DID_ANYTHING=y

    BUILD_DIR=build-$$

    # Copy header files
    echo "Building application in ./$BUILD_DIR"
    do_cmd mkdir "$BUILD_DIR"
    do_cmd cp Makefile.app "$BUILD_DIR"/Makefile
    pushd "$BUILD_DIR"
    echo "[+] Copying header files"
    do_cmd MINIOS_ROOT=../minios make include/mini-os
    popd

    # Compile application
    echo "[+] Compiling Go application in $APP_DIR"
    [ -z "$O_FLAG" ] && out="$BUILD_DIR"/a.out || out="$O_FLAG"
    out="$(realpath "$out")"
    include="$(realpath "$BUILD_DIR"/include)"
    goroot="$(realpath "../go")"

    opts=
    [ "$X_FLAG" = y ] && opts="$OPTS -x"
    [ "$A_FLAG" = y ] && opts="$OPTS -a"

    pushd "$APP_DIR"
    export GOPATH
    do_cmd GOROOT="$goroot" \
           CGO_ENABLED=1 \
           CGO_CFLAGS=-I"$include" \
           GOOS=netbsd \
           GOARCH=amd64 \
           "$goroot"/bin/go build -buildmode=c-archive $opts -o "$out"
    popd

    do_cmd rm -r "$BUILD_DIR"
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
