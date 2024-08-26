package barrow

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestParsePerm(t *testing.T) {
	i, err := strconv.ParseInt("0755", 8, 64)
	if err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%d %d\n", i, 0755)
}

func TestNameInArchive(t *testing.T) {
	prefix := "/usr/local"
	dest := "bin"
	baseName := "bali"
	nameInArchive := filepath.Join(prefix, dest, baseName)
	fmt.Fprintf(os.Stderr, "%s %s\n", nameInArchive, filepath.ToSlash(nameInArchive))
}
