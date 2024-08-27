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

# if ARGS=$(getopt -a -o p:hv --long prefix:,version,help -- "$@"); then
#     eval set -- "${ARGS}"
#     while :; do
#         case $1 in
#         -p | --prefix)
#             install_prefix=$2
#             install_prefix=$(fix_slashes "${install_prefix}")
#             shift
#             ;;
#         -h | --help)
#             usage
#             ;;
#         -v | --version)
#             echo "1.0.0"
#             exit 0
#             ;;
#         --)
#             shift
#             break
#             ;;
#         *)
#             echo "Internal error!"
#             exit 1
#             ;;
#         esac
#         shift
#     done
# else

# fi

for a in "$@"; do
    if echo "$a" | grep "^--prefix=" >/dev/null 2>/dev/null; then
        install_prefix="${a/--prefix=\///}"
        install_prefix=$(fix_slashes "${install_prefix}")
        continue
    fi
    if echo "$a" | grep "^--version" >/dev/null 2>/dev/null; then
        echo "1.0.0"
        exit 0
    fi
    if echo "$a" | grep "^--prefix" >/dev/null 2>/dev/null; then
        echo -e "error: \x1b[31m--prefix /path/to/prefix\x1b[0m is not support, switch to: \x1b[31m--prefix=/path/to/prefix\x1b[0m"
        exit 1
    fi
    if echo "$a" | grep "^--help" >/dev/null 2>/dev/null; then
        usage
    fi
done

prefix=$(pwd)
if [[ "x${install_prefix}" != "x" ]]; then
    prefix="${install_prefix}"
fi
package=$(basename "$0")
echo -e "The ${package} will be extracted to: \\x1b[32m${prefix}\\x1b[0m"
