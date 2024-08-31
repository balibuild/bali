#!/usr/bin/env bash

SCRIPT_FOLDER_REL=$(dirname "$0")
SCRIPT_FOLDER=$(
    cd "${SCRIPT_FOLDER_REL}" || exit
    pwd
)
TOPLEVEL_SOURCE_DIR=$(dirname "${SCRIPT_FOLDER}")
BALI_SOURCE_DIR="${TOPLEVEL_SOURCE_DIR}/cmd/bali"

if [[ "$OSTYPE" == "msys" ]]; then
    SUFFIX=".exe"
fi

echo -e "build root \x1b[32m${TOPLEVEL_SOURCE_DIR}\x1b[0m"

cd "$BALI_SOURCE_DIR" || exit 1
go build
cp "bali${SUFFIX}" "$TOPLEVEL_SOURCE_DIR/bali.exe"

cd "${TOPLEVEL_SOURCE_DIR}" || exit 1

case "$OSTYPE" in
solaris*)
    echo "solaris unsupported"
    ;;
darwin*)
    echo -e "bootstarp for \x1b[32mdarwin/amd64\x1b[0m"
    if ! "${TOPLEVEL_SOURCE_DIR}/bali.exe" '--pack=tar,sh' --target=darwin --arch=amd64; then
        echo "bootstrap bali failed"
        exit 1
    fi
    echo -e "bootstarp for \x1b[32mdarwin/arm64\x1b[0m"
    if ! "${TOPLEVEL_SOURCE_DIR}/bali.exe" '--pack=tar,sh' --target=darwin --arch=arm64; then
        echo "bootstrap bali failed"
        exit 1
    fi
    ;;
linux*)
    echo -e "bootstarp for \x1b[32mlinux/amd64\x1b[0m"
    if ! "${TOPLEVEL_SOURCE_DIR}/bali.exe" --pack='rpm,deb,tar,sh' --target=linux --arch=amd64; then
        echo "bootstrap bali failed"
        exit 1
    fi
    echo -e "bootstarp for \x1b[32mlinux/arm64\x1b[0m"
    if ! "${TOPLEVEL_SOURCE_DIR}/bali.exe" --pack='rpm,deb,tar,sh' --target=linux --arch=arm64; then
        echo "bootstrap bali failed"
        exit 1
    fi
    ;;
bsd*)
    echo "bsd unsupported"
    ;;
msys*)
    echo -e "bootstarp for \x1b[32mwindows/amd64\x1b[0m"
    if ! "${TOPLEVEL_SOURCE_DIR}/bali.exe" --pack=zip --target=windows --arch=amd64; then
        echo "bootstrap bali failed"
        exit 1
    fi
    echo -e "bootstarp for \x1b[32mwindows/arm64\x1b[0m"
    if ! "${TOPLEVEL_SOURCE_DIR}/bali.exe" --pack=zip --target=windows --arch=arm64; then
        echo "bootstrap bali failed"
        exit 1
    fi
    ;;
esac

echo -e "\\x1b[32mbali: bootstrap success\\x1b[0m"
