package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fcharlie/buna/debug/pe"
)

const (
	MaxDepth = 200
)

type Assets struct {
	files   []string
	depends map[string]string
	depth   int
}

func NewAssets(files []string) *Assets {
	return &Assets{files: files, depends: make(map[string]string)}
}

func (a *Assets) Parse() error {
	for _, f := range a.files {
		location := filepath.Dir(f)
		if err := a.parse(f, location); err != nil {
			return err
		}
	}
	return nil
}

func (a *Assets) unrecorded(dllname string) bool {
	if _, ok := a.depends[dllname]; ok {
		return false
	}
	return true
}

func (a *Assets) parse(filename, location string) error {
	a.depth++
	defer func() {
		a.depth--
	}()
	DbgPrint("current parse: %s", filepath.Base(filename))
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
		DbgPrint("find imports: %s", d)
		fixedLib := strings.ToLower(d)
		if !a.unrecorded(fixedLib) {
			continue // dll recorded
		}
		p := filepath.Join(location, d)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if a.depth > MaxDepth {
			continue
		}
		if err := a.parse(p, location); err != nil {
			return err
		}
		a.depends[fixedLib] = p
	}
	for d := range tables.Delay {
		DbgPrint("find delay imports: %s", d)
		fixedLib := strings.ToLower(d)
		if !a.unrecorded(fixedLib) {
			continue // dll recorded
		}
		p := filepath.Join(location, d)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if a.depth > MaxDepth {
			continue
		}
		if err := a.parse(p, location); err != nil {
			return err
		}
		a.depends[fixedLib] = p
	}
	return nil
}

func (a *Assets) Write(outfile string) error {
	fd, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer fd.Close()
	h := sha256.New()
	w := io.MultiWriter(fd, h)
	if err := a.compressFiles(w); err != nil {
		return err
	}
	hs := hex.EncodeToString(h.Sum(nil))
	fmt.Fprintf(os.Stderr, "\x1b[34m%s\x1b[0m: '\x1b[36mSHA256:%s\x1b[0m'", filepath.Base(outfile), hs)
	return nil
}

func (a *Assets) compressFiles(w io.Writer) error {
	zw := zip.NewWriter(w)
	for _, p := range a.depends {
		if err := a.compressFile(zw, p); err != nil {
			return err
		}
	}
	for _, f := range a.files {
		if err := a.compressFile(zw, f); err != nil {
			return err
		}
	}
	return zw.Close()
}

func (a *Assets) compressFile(zw *zip.Writer, filename string) error {
	fi, err := os.Stat(filename)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "peassets compress: \x1b[35m%s\x1b[0m\n", filepath.Base(filename))
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
