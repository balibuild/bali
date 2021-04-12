package pack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

// RespondName
const (
	RespondName = "bali_post_install.sh"
	header      = `#!/usr/bin/env bash
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
`
	respondheader = `#!/usr/bin/env bash
TOPLEVEL=$(dirname "$0")
bali_apply_target() {
    NEWNAME=$(basename "$1")
    DIRNAME=$(dirname "$1")
    NAME="${NEWNAME/%.new/}"
    TARGETFILE="$DIRNAME/$NAME"
    echo "apply target $TARGETFILE"
    if [[ ! -d "$DIRNAME/old" ]]; then
        mkdir -p "$DIRNAME/old"
    fi
    if [[ -f "$DIRNAME/old/$NAME.3" ]]; then
        rm "$DIRNAME/old/$NAME.3"
    fi
    if [[ -f "$DIRNAME/old/$NAME.2" ]]; then
        mv "$DIRNAME/old/$NAME.2" "$DIRNAME/old/$NAME.3"
    fi
    if [[ -f "$DIRNAME/old/$NAME.1" ]]; then
        mv "$DIRNAME/old/$NAME.1" "$DIRNAME/old/$NAME.2"
    fi
    if [[ -f "$DIRNAME/$NAME.old" ]]; then
        mv "$DIRNAME/$NAME.old" "$DIRNAME/old/$NAME.1"
    fi
    ###
    if [[ -f "$TARGETFILE" ]]; then
        mv "$TARGETFILE" "$TARGETFILE.old"
    fi
    mv "$TARGETFILE.new" "$TARGETFILE"
}
bali_apply_config() {
    NEWNAME=$(basename "$1")
    DIRNAME=$(dirname "$1")
    NAME="${NEWNAME/%.template/}"
    echo -e "install config \x1b[32m$DIRNAME/$NAME\x1b[0m"
    if [[ ! -d "$DIRNAME" ]]; then
        mkdir -p "$DIRNAME"
    fi
    if [[ -f "$DIRNAME/$NAME" ]]; then
        echo -e "File \x1b[33m$NAME\x1b[0m already exists in $DIRNAME"
        git --no-pager diff --no-index "$1" "$DIRNAME/$NAME"
        rm "$1"
    else
        echo -e "rename $1 to \x1b[32m$DIRNAME/$NAME\x1b[0m"
        mv "$1" "$DIRNAME/$NAME"
    fi
}
`
)

// HashableFile hash file
type HashableFile struct {
	fd *os.File
	h  hash.Hash
	mw io.Writer
}

// WriteString write string
func (f *HashableFile) WriteString(s string) (int, error) {
	return f.mw.Write([]byte(s))
}

// Close close file
func (f *HashableFile) Close() error {
	if f == nil || f.fd == nil {
		return nil
	}
	return f.fd.Close()
}

// Hashsum hash sum
func (f *HashableFile) Hashsum(name string) {
	if f == nil {
		return
	}
	if f.h == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "\x1b[34m%s  %s\x1b[0m\n", hex.EncodeToString(f.h.Sum(nil)), name)
}

// Write a file
func (f *HashableFile) Write(p []byte) (int, error) {
	return f.mw.Write(p)
}

// OpenHashableFile todo
func OpenHashableFile(name string) (*HashableFile, error) {
	fd, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return nil, err
	}
	file := &HashableFile{h: sha256.New(), fd: fd}
	file.mw = io.MultiWriter(file.fd, file.h)
	if _, err := file.WriteString(header); err != nil {
		_ = file.Close()
		_ = os.Remove(name)
		return nil, err
	}
	return file, nil
}

// RespondWriter todo
type RespondWriter struct {
	fd   *os.File
	Path string
}

// Initialize todo
func (rw *RespondWriter) Initialize() error {
	rw.Path = filepath.Join(os.TempDir(), RespondName)
	var err error
	if rw.fd, err = os.OpenFile(rw.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755); err != nil {
		return err
	}
	return nil
}

// WriteBase todo
func (rw *RespondWriter) WriteBase() error {
	if rw.fd == nil {
		return nil
	}
	if _, err := rw.fd.WriteString(respondheader); err != nil {
		return err
	}
	return nil
}

// AddTarget todo
func (rw *RespondWriter) AddTarget(relname string) error {
	if rw.fd == nil {
		return nil
	}
	relname = filepath.ToSlash(relname)
	fmt.Fprintf(rw.fd, "echo -e \"install target \\x1b[32m$TOPLEVEL/%s\\x1b[0m\"\nbali_apply_target \"$TOPLEVEL/%s\"\n", relname, relname)
	return nil
}

// AddProfile todo
func (rw *RespondWriter) AddProfile(relname string) error {
	if rw.fd == nil {
		return nil
	}
	relname = filepath.ToSlash(relname)
	fmt.Fprintf(rw.fd, "echo -e \"apply config \\x1b[32m$TOPLEVEL/%s\\x1b[0m\"\nbali_apply_config \"$TOPLEVEL/%s\"\n", relname, relname)
	return nil
}

// Close todo
func (rw *RespondWriter) Close() error {
	if rw.fd == nil {
		return nil
	}
	fmt.Fprintf(rw.fd, "rm -f \"$TOPLEVEL/%s\"\n", RespondName)
	return rw.fd.Close()
}
