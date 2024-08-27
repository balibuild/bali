#!/usr/bin/env bash

usage() {
    cat <<EOF
Usage: $0 [options]
Options: [defaults in brackets after descriptions]
  --help            print this message
  --version         print bali installer version
  --prefix=dir      directory in which to install
EOF
    exit 1
}

fix_slashes() {
    echo "$1" | sed 's/\\/\//g'
}

if ARGS=$(getopt -a -o p:hv --long prefix:,version,help -- "$@"); then
    eval set -- "${ARGS}"
    while :; do
        case $1 in
        -p | --prefix)
            install_prefix=$2
            install_prefix=$(fix_slashes "${install_prefix}")
            shift
            ;;
        -h | --help)
            usage
            ;;
        -v | --version)
            echo "1.0.0"
            exit 0
            ;;
        --)
            shift
            break
            ;;
        *)
            echo "Internal error!"
            exit 1
            ;;
        esac
        shift
    done
else
    for a in "$@"; do
        if echo "$a" | grep "^--prefix=" >/dev/null 2>/dev/null; then
            install_prefix="${a/--prefix=\///}"
            install_prefix=$(fix_slashes "${install_prefix}")
            continue
        fi
        if echo "$a" | grep "^--prefix" >/dev/null 2>/dev/null; then
            echo -e "error: \x1b[31m--prefix /path/to/prefix\x1b[0m is not support, switch to: \x1b[31m--prefix=/path/to/prefix\x1b[0m"
            exit 1
        fi
        if echo "$a" | grep "^--version" >/dev/null 2>/dev/null; then
            echo "1.0.0"
            exit 0
        fi
        if echo "$a" | grep "^--help" >/dev/null 2>/dev/null; then
            usage
        fi
    done
fi

echo "This is a self-extracting archive."
prefix=$(pwd)
if [[ "x${install_prefix}" != "x" ]]; then
    prefix="${install_prefix}"
fi
package=$(basename "$0")
echo -e "The ${package} will be extracted to: \\x1b[32m${prefix}\\x1b[0m"
if [ ! -d "${prefix}" ]; then
    mkdir -p "${prefix}" || exit 1
fi
echo
echo "Using traget directory: ${prefix}"
echo "Extracting, please wait..."
echo ""
ARCHIVE=$(awk '/^__ARCHIVE_BELOW__/ {print NR + 1; exit 0; }' "$0")
tail "-n+$ARCHIVE" "$0" | tar xzvm -C "$prefix" >/dev/null 2>&1 3>&1
if [[ -f "${prefix}/post-install.sh" ]]; then
    chmod +x "${prefix}/post-install.sh"
    echo -e "\\x1b[33mrun ${prefix}/post-install.sh\\x1b[0m"
    bash "${prefix}/post-install.sh"
fi
exit 0
#This line must be the last line of the file
# shellcheck disable=SC2317
__ARCHIVE_BELOW__
