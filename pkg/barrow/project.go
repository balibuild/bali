package barrow

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type File struct {
	Path        string `toml:"path"`
	Destination string `toml:"destination"`
	Rename      string `toml:"rename,omitempty"`      // when rename exists: rename file to some name
	Permissions string `toml:"permissions,omitempty"` // 0755 0644
}

type Project struct {
	Name        string `toml:"name"`
	Summary     string `toml:"summary,omitempty"`     // Is a short description of the software
	Description string `toml:"description,omitempty"` // description is a longer piece of software information than Summary, consisting of one or more paragraphs
	Version     string `toml:"version,omitempty"`
	Vendor      string `toml:"vendor,omitempty"`
	URL         string `toml:"url,omitempty"`
	Packager    string `toml:"url,omitempty"` // BALI_RPM_PACKAGER
	Group       string `toml:"group,omitempty"`
	License     string `toml:"license,omitempty"`
	Prefix      string `toml:"prefix,omitempty"` // install prefix: rpm required
	Targets     string `toml:"targets,omitempty"`
	Files       string `toml:"files,omitempty"`
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

func NewProject(cwd string) (*Project, error) {
	file := filepath.Join(cwd, "bali.toml")
	var p Project
	if err := LoadMetadata(file, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
