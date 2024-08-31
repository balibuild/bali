package barrow

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func nonEmpty(a string, dv string) string {
	if len(a) != 0 {
		return a
	}
	return dv
}

func isSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

// nopCloser wrap io.Writer --> io.WriteCloser
type nopCloser struct {
	io.Writer
}

func (w nopCloser) Close() error {
	return nil
}

// ToNixPath, AsExplicitRelativePath, AsRelativePath (MIT License)
// Thanks: https://github.com/goreleaser/nfpm/blob/main/files/files.go

// ToNixPath converts the given path to a nix-style path.
//
// Windows-style path separators are considered escape
// characters by some libraries, which can cause issues.
func ToNixPath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

// As relative path converts a path to an explicitly relative path starting with
// a dot (e.g. it converts /foo -> ./foo and foo -> ./foo).
func AsExplicitRelativePath(path string) string {
	return "./" + AsRelativePath(path)
}

// AsRelativePath converts a path to a relative path without a "./" prefix. This
// function leaves trailing slashes to indicate that the path refers to a
// directory, and converts the path to Unix path.
func AsRelativePath(path string) string {
	cleanedPath := strings.TrimLeft(ToNixPath(path), "/")
	if len(cleanedPath) > 1 && strings.HasSuffix(path, "/") {
		return cleanedPath + "/"
	}
	return cleanedPath
}

// NormalizeAbsoluteFilePath returns an absolute cleaned path separated by
// slashes.
func NormalizeAbsoluteFilePath(src string) string {
	return ToNixPath(filepath.Join("/", src))
}

// normalizeFirPath is linke NormalizeAbsoluteFilePath with a trailing slash.
func NormalizeAbsoluteDirPath(path string) string {
	return NormalizeAbsoluteFilePath(strings.TrimRight(path, "/")) + "/"
}

func hashPrint(h hash.Hash, name string) {
	fmt.Fprintf(os.Stdout, "\x1b[38;2;0;191;255m%s  %s\x1b[0m\n", hex.EncodeToString(h.Sum(nil)), name)
}

func stage(s string, format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[\x1b[38;2;63;247;166m%s\x1b[0m] \x1b[38;02;39;199;173m%s\x1b[0m\n", s, fmt.Sprintf(format, a...))
}

func status(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "$> \x1b[38;02;245;202;100m%s\x1b[0m\n", fmt.Sprintf(format, a...))
}

// func status(format string, a ...any) {
// 	var style = lipgloss.NewStyle().
// 		Bold(true).
// 		Foreground(lipgloss.Color("#F5CA64")).
// 		PaddingTop(0).
// 		PaddingLeft(1)

// 	fmt.Println(style.Render(fmt.Sprintf(format, a...)))
// }

func cmdStringsArgs(c *exec.Cmd) string {
	b := new(strings.Builder)
	b.WriteString("go")
	for _, a := range c.Args[1:] {
		b.WriteByte(' ')
		b.WriteString(a)
	}
	return b.String()
}
