package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/balibuild/bali/v2/base"
	"github.com/pelletier/go-toml"
)

// Executable todo
type Executable struct {
	Name             string   `toml:"name"`
	Destination      string   `toml:"destination,omitempty"`
	Description      string   `toml:"description,omitempty"`
	Version          string   `toml:"version,omitempty"`
	Links            []string `toml:"links,omitempty"` // create symlink
	GoFlags          []string `toml:"goflags,omitempty"`
	VersionInfo      string   `toml:"versioninfo,omitempty"`
	IconPath         string   `toml:"icon,omitempty"`
	Manifest         string   `toml:"manifest,omitempty"`
	BuildConstraints string   `toml:"build,omitempty"` // Build Constraints
}

// IsMeetsConstraints todo
func (e *Executable) IsMeetsConstraints(target, arch string) bool {
	if len(e.BuildConstraints) == 0 {
		return true
	}
	// TODO support golang style build constraints
	return true
}

// File todo
type File struct {
	Path        string `toml:"path"`
	Destination string `toml:"destination"`
	NewName     string `toml:"newname,omitempty"`
	NoRename    bool   `toml:"norename,omitempty"`
	Executable  bool   `toml:"executable,omitempty"`
}

// Base get BaliFile base
func (file *File) Base() string {
	if len(file.NewName) != 0 {
		return file.NewName
	}
	return filepath.Base(file.Path)
}

// Configure configure to out dir
func (file *File) Configure(workdir, outdir string) error {
	fileoutdir := filepath.Join(outdir, file.Destination)
	_ = os.MkdirAll(fileoutdir, 0775)
	var outfile string
	if len(file.NewName) != 0 {
		outfile = filepath.Join(fileoutdir, file.NewName)
	} else {
		name := filepath.Base(file.Path)
		outfile = filepath.Join(fileoutdir, name)
	}
	srcfile := filepath.Join(workdir, file.Path)
	if base.PathExists(outfile) && !IsForceMode {
		if !IsForceMode {
			return nil
		}
		fmt.Fprintf(os.Stderr, "update \x1b[32m%s\x1b[0m\n", outfile)
	} else {
		fmt.Fprintf(os.Stderr, "install \x1b[32m%s\x1b[0m\n", outfile)
	}
	return base.CopyFile(srcfile, outfile)
}

// Project  todo
type Project struct {
	Name        string   `toml:"name"`
	Version     string   `toml:"version,omitempty"`
	Destination string   `toml:"destination,omitempty"`
	Files       []File   `toml:"files,omitempty"`
	Dirs        []string `toml:"dirs,omitempty"`
	Respond     string   `toml:"respond,omitempty"`
}

// FileConfigure todo
func (bm *Project) FileConfigure(workdir, outdir string) error {
	for _, file := range bm.Files {
		if err := file.Configure(workdir, outdir); err != nil {
			return err
		}
	}
	return nil
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
