#!/usr/bin/env bash

function Xrealpath() {
    local success=true
    local path="$1"

    # make sure the string isn't empty as that implies something in further logic
    if [ -z "$path" ]; then
        success=false
    else
        # start with the file name (sans the trailing slash)
        path="${path%/}"

        # if we stripped off the trailing slash and were left with nothing, that means we're in the root directory
        if [ -z "$path" ]; then
            path="/"
        fi

        # get the basename of the file (ignoring '.' & '..', because they're really part of the path)
        local file_basename="${path##*/}"
        if [[ ("$file_basename" = ".") || ("$file_basename" = "..") ]]; then
            file_basename=""
        fi

        # extracts the directory component of the full path, if it's empty then assume '.' (the current working directory)
        local directory="${path%$file_basename}"
        if [ -z "$directory" ]; then
            directory='.'
        fi

        # attempt to change to the directory
        if ! cd "$directory" &>/dev/null; then
            success=false
        fi

        if $success; then
            # does the filename exist?
            if [[ (-n "$file_basename") && (! -e "$file_basename") ]]; then
                success=false
            fi

            # get the absolute path of the current directory & change back to previous directory
            local abs_path
            abs_path="$(pwd -P)"
            cd "-" &>/dev/null || exit 1

            # Append base filename to absolute path
            if [ "${abs_path}" = "/" ]; then
                abs_path="${abs_path}${file_basename}"
            else
                abs_path="${abs_path}/${file_basename}"
            fi

            # output the absolute path
            echo "$abs_path"
        fi
    fi

    $success
}

REALPATH=$(Xrealpath "$0")
SCRIPTROOT=$(dirname "$REALPATH")
TOPLEVEL=$(dirname "$SCRIPTROOT")
SRCDIR="${TOPLEVEL}/cmd/bali"

echo -e "build root \\x1b[32m${TOPLEVEL}\\x1b[0m"
OLDPWD=$(pwd)
cd "$SRCDIR" || exit 1
go build
cp "bali" "$TOPLEVEL/bali.out"

cd "${TOPLEVEL}" || exit 1
if ! "./bali.out" -z; then
    echo "bootstrap bali failed"
    exit 1
fi
echo -e "\\x1b[32mbootstrap bali success\\x1b[0m"
