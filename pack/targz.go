package pack

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/balibuild/bali/utilities"
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

// Insert todo
func (pk *TargzPacker) Insert(src, nameInArchive string) error {
	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	var linkTarget string
	if isSymlink(st) {
		if linkTarget, err = os.Readlink(src); err != nil {
			return utilities.ErrorCat(src, ": readlink: ", err.Error())
		}
	}
	hdr, err := tar.FileInfoHeader(st, linkTarget)
	if err != nil {
		return utilities.ErrorCat(src, ": marking header: ", err.Error())
	}
	hdr.Name = filepath.ToSlash(nameInArchive)
	if err = pk.tw.WriteHeader(hdr); err != nil {
		return utilities.ErrorCat(nameInArchive, ": write header:", err.Error())
	}
	if st.IsDir() {
		return nil
	}
	if hdr.Typeflag != tar.TypeReg {
		return nil
	}
	fd, err := os.Open(src)
	if err != nil {
		return utilities.ErrorCat(src, ": opening: ", err.Error())
	}
	defer fd.Close()
	if _, err := io.Copy(pk.tw, fd); err != nil {
		return utilities.ErrorCat(src, ": copying contents: ", err.Error())
	}
	return nil
}
