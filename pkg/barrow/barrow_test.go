package barrow

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestParsePerm(t *testing.T) {
	i, err := strconv.ParseInt("0755", 8, 64)
	if err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%d %d %d\n", i, 0755, 0o755)
}

func TestNameInArchive(t *testing.T) {
	prefix := "/usr/local"
	dest := "bin"
	baseName := "bali"
	nameInArchive := filepath.Join(prefix, dest, baseName)
	fmt.Fprintf(os.Stderr, "%s %s\n", nameInArchive, ToNixPath(nameInArchive))
}

func TestEncodePackage(t *testing.T) {
	p := &Package{
		Name: "jack",
		Include: []*FileItem{
			{
				Path:        "LICENSE",
				Destination: "share",
			},
			{
				Path:        "README.md",
				Destination: "/usr/share",
			},
		},
	}
	if err := toml.NewEncoder(os.Stderr).Encode(p); err != nil {
		fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
	}
}
