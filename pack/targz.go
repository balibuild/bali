package pack

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/balibuild/bali/v2/base"
)

// ISVTX todo
const (
	TarISVTX = 01000
)

// TargzPacker todo
type TargzPacker struct {
	tw *tar.Writer
	gw *gzip.Writer
}

// NewTargzPacker todo
func NewTargzPacker(w io.Writer) *TargzPacker {
	pk := &TargzPacker{gw: gzip.NewWriter(w)}
	pk.tw = tar.NewWriter(pk.gw)
	return pk
}

// Close todo
func (pk *TargzPacker) Close() error {
	if pk.tw != nil {
		pk.tw.Close()
	}
	if pk.gw != nil {
		return pk.gw.Close()
	}
	return nil
}

// AddTargetLink todo
func (pk *TargzPacker) AddTargetLink(nameInArchive, linkName string) error {
	hdr := &tar.Header{
		Name:     filepath.ToSlash(nameInArchive),
		ModTime:  time.Now(),
		Mode:     0755,
		Typeflag: tar.TypeSymlink,
		Linkname: filepath.ToSlash(linkName)}
	if err := pk.tw.WriteHeader(hdr); err != nil {
		return base.ErrorCat(nameInArchive, ": write header:", err.Error())
	}
	return nil
}

// AddFileEx todo
func (pk *TargzPacker) AddFileEx(src, nameInArchive string, exerights bool) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	var linkTarget string
	if isSymlink(fi) {
		if linkTarget, err = os.Readlink(src); err != nil {
			return base.ErrorCat(src, ": readlink: ", err.Error())
		}
	}
	hdr, err := tar.FileInfoHeader(fi, linkTarget)
	if err != nil {
		return base.ErrorCat(src, ": marking header: ", err.Error())
	}
	if exerights {
		hdr.Mode = hdr.Mode | 0755
	}
	hdr.Name = filepath.ToSlash(nameInArchive)
	if err = pk.tw.WriteHeader(hdr); err != nil {
		return base.ErrorCat(nameInArchive, ": write header:", err.Error())
	}
	if fi.IsDir() {
		return nil
	}
	if hdr.Typeflag != tar.TypeReg {
		return nil
	}
	fd, err := os.Open(src)
	if err != nil {
		return base.ErrorCat(src, ": opening: ", err.Error())
	}
	defer fd.Close()
	if _, err := io.Copy(pk.tw, fd); err != nil {
		return base.ErrorCat(src, ": copying contents: ", err.Error())
	}
	return nil
}

// AddFile todo
func (pk *TargzPacker) AddFile(src, nameInArchive string) error {
	return pk.AddFileEx(src, nameInArchive, false)
}
