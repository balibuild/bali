package barrow

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type FileItem struct {
	Path        string `toml:"path"`
	Destination string `toml:"destination"`
	Rename      string `toml:"rename,omitempty"`      // when rename is no empty: rename file to name
	Permissions string `toml:"permissions,omitempty"` // 0755 0644
}

type Package struct {
	Name        string     `toml:"name"`
	Summary     string     `toml:"summary,omitempty"`     // Is a short description of the software
	Description string     `toml:"description,omitempty"` // description is a longer piece of software information than Summary, consisting of one or more paragraphs
	Version     string     `toml:"version,omitempty"`
	Authors     []string   `toml:"authors,omitempty"`
	Vendor      string     `toml:"vendor,omitempty"`
	URL         string     `toml:"url,omitempty"`
	Packager    string     `toml:"url,omitempty"` // BALI_RPM_PACKAGER
	Group       string     `toml:"group,omitempty"`
	License     string     `toml:"license,omitempty"`
	LicenseFile string     `toml:"license-file,omitempty"`
	Prefix      string     `toml:"prefix,omitempty"` // install prefix: rpm required
	Crates      []string   `toml:"crates,omitempty"`
	Include     []FileItem `toml:"include,omitempty"`
}

func LoadMetadata(file string, v any) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	if err := toml.NewDecoder(fd).Decode(v); err != nil {
		return err
	}
	return nil
}

func LoadPackage(cwd string) (*Package, error) {
	file := filepath.Join(cwd, "bali.toml")
	var p Package
	if err := LoadMetadata(file, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
