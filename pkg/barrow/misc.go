package barrow

import (
	"io"
	"os"
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
