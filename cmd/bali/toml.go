package main

import (
	"os"

	"github.com/pelletier/go-toml"
)

// ExecutableEx core
type ExecutableEx struct {
	Name        string   `toml:"name"`
	Destination string   `toml:"destination,omitempty"`
	Description string   `toml:"description,omitempty"`
	Version     string   `toml:"version,omitempty"`
	Links       []string `toml:"links,omitempty"` // create symlink
	GoFlags     []string `toml:"goflags,omitempty"`
	VersionInfo string   `toml:"versioninfo,omitempty"`
	IconPath    string   `toml:"icon,omitempty"`
	Manifest    string   `toml:"manifest,omitempty"`
}

// FileEx todo
type FileEx struct {
	Path        string `toml:"path"`
	Destination string `toml:"destination"`
	NewName     string `toml:"newname,omitempty"`
	NoRename    bool   `toml:"norename,omitempty"`
	Executable  bool   `toml:"executable,omitempty"`
}

// ProjectEx  todo
type ProjectEx struct {
	Name    string   `toml:"name"`
	Version string   `toml:"version,omitempty"`
	Files   []FileEx `toml:"files,omitempty"`
	Dirs    []string `toml:"dirs,omitempty"`
	Respond string   `toml:"respond,omitempty"`
}

// LoadTomlMetadata todo
func LoadTomlMetadata(file string, v interface{}) error {
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
