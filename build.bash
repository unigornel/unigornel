#!/bin/bash

set -e

function usage {
    echo "usage: build.bash help|-h|--help"
    echo "       build.bash go --gobootstrap path"
    echo "       build.bash minios --goarchive app/app.a --goinclude app/"
    echo "       build.bash compile --app path/to/app/dir/ [-o path/to/app.a] [-a] [-x]"
    echo "       build.bash app --app path/to/app/dir/ [-o minios] [-a] [-x]"
}


HELP=n
[ -z "$GOROOT_BOOTSTRAP" ] && GOROOT_BOOTSTRAP=
MINIOS_GOARCHIVE=
MINIOS_GOINCLUDE=
APP_DIR=
O_FLAG=
A_FLAG=n
X_FLAG=n

SCRIPT="$0"
CMD="$1"
shift

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
if [ "$HELP" = y -o "$CMD" = help ]; then
    usage
    exit 0
fi

# Build Go
if [ "$CMD" = go ]; then
    if [ -n "$GOROOT_BOOTSTRAP" ]; then
        pushd go/src
        echo "Building Go in $PWD"
        do_cmd GOROOT_BOOTSTRAP="$GOROOT_BOOTSTRAP" GOOS=netbsd GOARCH=amd64 ./make.bash
        popd
    else
        error "missing --gobootstrap flag"
    fi

# Compile Go
elif [ "$CMD" = "compile" ]; then
    if [ -n "$APP_DIR" ]; then
        [ -z "$GOPATH" ] && error "cannot build app: GOPATH not set"

        BUILD_DIR=build-$$

        # Copy header files
        echo "Compiling Go in ./$BUILD_DIR"
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
        [ "$X_FLAG" = y ] && opts="$opts -x"
        [ "$A_FLAG" = y ] && opts="$opts -a"

        pushd "$APP_DIR"
        export GOPATH
        do_cmd GOROOT="$goroot" \
               CGO_ENABLED=1 \
               CGO_CFLAGS=-I"$include" \
               GOOS=netbsd \
               GOARCH=amd64 \
               "$goroot"/bin/go build -buildmode=c-archive $opts -o "$out"
        do_cmd objcopy --globalize-symbol=_rt0_amd64_netbsd_lib "$out"
        popd

        do_cmd rm -r "$BUILD_DIR"
    else
        error "missing --app flag"
    fi

# Build Mini-OS
elif [ "$CMD" = minios ]; then
    if [ -n "$MINIOS_GOARCHIVE" -a -n "$MINIOS_GOINCLUDE" ]; then
        archive="$(realpath "$MINIOS_GOARCHIVE")"
        include="$(realpath "$MINIOS_GOINCLUDE")"

        pushd minios
        echo "Building Mini-OS in $PWD"
        do_cmd make GOARCHIVE="$archive" GOINCLUDE="$include"
        popd

    else
        error "missing --goarchive and/or --goinclude flag"
    fi

# Build full application
elif [ "$CMD" = app ]; then
    if [ -n "$APP_DIR" ]; then
        BUILD_DIR=build-$$
        echo "Building application in ./$BUILD_DIR"
        do_cmd mkdir "$BUILD_DIR"

        # Compile Go
        echo "[+] Compiling Go code"
        export GOPATH
        opts=
        [ "$X_FLAG" = y ] && opts="$opts -x"
        [ "$A_FLAG" = y ] && opts="$opts -a"
        do_cmd "$SCRIPT" compile --app "$APP_DIR" $opts -o "$BUILD_DIR"/app.a

        # Compile Mini-OS
        echo "[+] Compiling Mini-OS"
        do_cmd "$SCRIPT" minios --goarchive "$BUILD_DIR"/app.a --goinclude "$BUILD_DIR"
        do_cmd rm -r "$BUILD_DIR"

        # Copy minios
        if [ -n "$O_FLAG" ]; then
            do_cmd cp minios/mini-os "$O_FLAG"
            echo "[+] Unikernel is in $O_FLAG"
        else
            echo "[+] Unikernel is in minios/mini-os"
        fi
    else
        error "missing --app flag"
    fi

# Unknown command
else
    usage
    error "unknown command: $CMD"
fi