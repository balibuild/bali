package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fcharlie/buna/debug/pe"
)

const (
	MaxDepth = 200
)

type Assets struct {
	filename string
	location string
	depends  map[string]string
	depth    int
}

func NewAssets(filename string) *Assets {
	return &Assets{filename: filename, location: filepath.Dir(filename), depends: make(map[string]string)}
}

func (a *Assets) Parse() error {
	return a.parse(a.filename)
}

func (a *Assets) isunrecorded(dllname string) bool {
	if _, ok := a.depends[dllname]; ok {
		return false
	}
	return true
}

func (a *Assets) parse(filename string) error {
	a.depth++
	defer func() {
		a.depth--
	}()
	fd, err := pe.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()
	tables, err := fd.LookupFunctionTable()
	if err != nil {
		return err
	}
	for d := range tables.Imports {
		if !a.isunrecorded(d) {
			continue // dll recorded
		}
		p := filepath.Join(a.location, d)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if a.depth > MaxDepth {
			continue
		}
		if err := a.parse(p); err != nil {
			return err
		}
		a.depends[d] = p
	}
	for d := range tables.Delay {
		if !a.isunrecorded(d) {
			continue // dll recorded
		}
		p := filepath.Join(a.location, d)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if a.depth > MaxDepth {
			continue
		}
		if err := a.parse(p); err != nil {
			return err
		}
		a.depends[d] = p
	}
	return nil
}

func (a *Assets) Write(outfile string) error {
	fd, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer fd.Close()
	zw := zip.NewWriter(fd)
	for _, p := range a.depends {
		if err := a.compressFile(zw, p); err != nil {
			return err
		}
	}
	if err := a.compressFile(zw, a.filename); err != nil {
		return err
	}
	return zw.Close()
}

func (a *Assets) compressFile(zw *zip.Writer, filename string) error {
	fi, err := os.Stat(filename)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "compress %s\n", filename)
	hdr, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return nil
	}
	hdr.Name = filepath.Base(filename)
	hdr.Method = zip.Deflate
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()
	if _, err := io.Copy(w, fd); err != nil {
		return err
	}
	return nil
}
