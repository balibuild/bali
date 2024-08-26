#!/usr/bin/env bash
# Display usage
bali_usage() {
    cat <<EOF
Usage: $0 [options]
Options: [defaults in brackets after descriptions]
  --help            print this message
  --version         print cmake installer version
  --prefix=dir      directory in which to install
EOF
    exit 1
}
bali_fix_slashes() {
    echo "$1" | sed 's/\\/\//g'
}
bali_echo_exit() {
    echo "$1"
    exit 1
}
for a in "$@"; do
    if echo "$a" | grep "^--prefix=" >/dev/null 2>/dev/null; then
        bali_prefix_dir="${a/--prefix=\///}"
        bali_prefix_dir=$(bali_fix_slashes "${bali_prefix_dir}")
    fi
    if echo "$a" | grep "^--help" >/dev/null 2>/dev/null; then
        bali_usage
    fi
done
echo "This is a self-extracting archive."
toplevel=$(pwd)
if [[ "x${bali_prefix_dir}" != "x" ]]; then
    toplevel="${bali_prefix_dir}"
fi
package=$(basename "$0")
echo -e "The ${package} will be extracted to: \\x1b[32m${toplevel}\\x1b[0m"
if [ ! -d "${toplevel}" ]; then
    mkdir -p "${toplevel}" || exit 1
fi
echo
echo "Using traget directory: ${toplevel}"
echo "Extracting, please wait..."
echo ""
ARCHIVE=$(awk '/^__ARCHIVE_BELOW__/ {print NR + 1; exit 0; }' "$0")
tail "-n+$ARCHIVE" "$0" | tar xzvm -C "$toplevel" >/dev/null 2>&1 3>&1
if [[ -f "${toplevel}/bali_post_install.sh" ]]; then
	chmod +x "${toplevel}/bali_post_install.sh"
	echo -e "\\x1b[33mrun ${toplevel}/bali_post_install.sh\\x1b[0m"
    bash "${toplevel}/bali_post_install.sh"
fi
exit 0
#This line must be the last line of the file
__ARCHIVE_BELOW__
