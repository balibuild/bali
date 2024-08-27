#!/usr/bin/env bash

SCRIPT_FOLDER_REL=$(dirname "$0")
SCRIPT_FOLDER=$(
    cd "${SCRIPT_FOLDER_REL}" || exit
    pwd
)
SOURCE_DIR=$(dirname "${SCRIPT_FOLDER}")
BALI_SOURCE_DIR="${SOURCE_DIR}/cmd/bali"

if [[ "$OSTYPE" == "msys" ]]; then
    SUFFIX=".exe"
fi

echo -e "build root \\x1b[32m${SOURCE_DIR}\\x1b[0m"

cd "$BALI_SOURCE_DIR" || exit 1
go build
cp "bali${SUFFIX}" "$SOURCE_DIR/bali${SUFFIX}"

cd "${SOURCE_DIR}" || exit 1
if ! "${SOURCE_DIR}/bali${SUFFIX}" --pack=tar; then
    echo "bootstrap bali failed"
    exit 1
fi
echo -e "\\x1b[32mbootstrap bali success\\x1b[0m"
