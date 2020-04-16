package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/balibuild/bali/utilities"
)

// Executable todo
type Executable struct {
	Name        string   `json:"name"`
	Destination string   `json:"destination,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Links       []string `json:"links,omitempty"` // create symlink
	GoFlags     []string `json:"goflags,omitempty"`
}

// File todo
type File struct {
	Path        string `json:"path"`
	Destination string `json:"destination"`
	NewName     string `json:"newname,omitempty"`
	NoRename    bool   `json:"norename,omitempty"`
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
	if utilities.PathExists(outfile) && !IsForceMode {
		if !IsForceMode {
			return nil
		}
		fmt.Fprintf(os.Stderr, "update \x1b[32m%s\x1b[0m\n", outfile)
	} else {
		fmt.Fprintf(os.Stderr, "install \x1b[32m%s\x1b[0m\n", outfile)
	}
	return utilities.CopyFile(srcfile, outfile)
}

// Project  todo
type Project struct {
	Name    string   `json:"name"`
	Version string   `json:"version,omitempty"`
	Files   []File   `json:"files,omitempty"`
	Dirs    []string `json:"dirs,omitempty"`
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

// LoadMetadata todo
func LoadMetadata(file string, v interface{}) error {
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
