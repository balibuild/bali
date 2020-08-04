package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/balibuild/bali/base"
	"github.com/pelletier/go-toml"
)

// Executable todo
type Executable struct {
	Name        string   `json:"name" toml:"name"`
	Destination string   `json:"destination,omitempty" toml:"destination,omitempty"`
	Description string   `json:"description,omitempty" toml:"description,omitempty"`
	Version     string   `json:"version,omitempty" toml:"version,omitempty"`
	Links       []string `json:"links,omitempty" toml:"links,omitempty"` // create symlink
	GoFlags     []string `json:"goflags,omitempty" toml:"goflags,omitempty"`
	VersionInfo string   `json:"versioninfo,omitempty" toml:"versioninfo,omitempty"`
	IconPath    string   `json:"icon,omitempty" toml:"icon,omitempty"`
	Manifest    string   `json:"manifest,omitempty" toml:"manifest,omitempty"`
}

// File todo
type File struct {
	Path        string `json:"path" toml:"path"`
	Destination string `json:"destination" toml:"destination"`
	NewName     string `json:"newname,omitempty" toml:"newname,omitempty"`
	NoRename    bool   `json:"norename,omitempty" toml:"norename,omitempty"`
	Executable  bool   `json:"executable,omitempty" toml:"executable,omitempty"`
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
	Name    string   `json:"name" toml:"name"`
	Version string   `json:"version,omitempty" toml:"version,omitempty"`
	Files   []File   `json:"files,omitempty" toml:"files,omitempty"`
	Dirs    []string `json:"dirs,omitempty" toml:"dirs,omitempty"`
	Respond string   `json:"respond,omitempty" toml:"respond,omitempty"`
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

// LoadJSONMetadata todo
func LoadJSONMetadata(file string, v interface{}) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	if err := json.NewDecoder(fd).Decode(v); err != nil {
		return err
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
