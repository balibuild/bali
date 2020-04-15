package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/balibuild/bali/utilities"
)

// BaliSrcMetadata todo
type BaliSrcMetadata struct {
	Name        string   `json:"name"`
	Destination string   `json:"destination,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	GoFlags     []string `json:"goflags,omitempty"`
}

// BaliFile todo
type BaliFile struct {
	Path        string `json:"path"`
	Destination string `json:"destination"`
	Rename      string `json:"rename,omitempty"`
}

// Base get BaliFile base
func (file *BaliFile) Base() string {
	if len(file.Rename) != 0 {
		return file.Rename
	}
	return filepath.Base(file.Path)
}

// Configure configure to out dir
func (file *BaliFile) Configure(workdir, outdir string) error {
	fileoutdir := filepath.Join(outdir, file.Destination)
	_ = os.MkdirAll(fileoutdir, 0775)
	var outfile string
	if len(file.Rename) != 0 {
		outfile = filepath.Join(fileoutdir, file.Rename)
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

// BaliMetadata  todo
type BaliMetadata struct {
	Name    string     `json:"name"`
	Version string     `json:"version,omitempty"`
	Files   []BaliFile `json:"files,omitempty"`
	Dirs    []string   `json:"dirs,omitempty"`
}

// FileConfigure todo
func (bm *BaliMetadata) FileConfigure(workdir, outdir string) error {
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
