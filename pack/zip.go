package pack

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/balibuild/bali/utilities"
)

// ZipPacker todo
type ZipPacker struct {
	zw *zip.Writer
}

// NewZipPacker todo
func NewZipPacker(w io.Writer) *ZipPacker {
	return &ZipPacker{zw: zip.NewWriter(w)}
}

// Close todo
func (zp *ZipPacker) Close() error {
	if zp.zw == nil {
		return nil
	}
	return zp.zw.Close()
}

// Insert file to zip packer
func (zp *ZipPacker) Insert(src, nameInArchive string) error {
	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(st)
	if err != nil {
		return utilities.ErrorCat(src, ": getting header: ", err.Error())
	}
	if st.IsDir() {
		header.Name = utilities.StrCat(nameInArchive, "/")
		header.Method = zip.Store
	} else {
		header.Name = nameInArchive
		header.Method = zip.Deflate
	}
	writer, err := zp.zw.CreateHeader(header)
	if err != nil {
		return utilities.ErrorCat(nameInArchive, ": making header:", err.Error())
	}
	if st.IsDir() {
		return nil
	}
	if isSymlink(st) {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return utilities.ErrorCat(src, ": readlink: ", err.Error())
		}
		if _, err := writer.Write([]byte(filepath.ToSlash(linkTarget))); err != nil {
			return utilities.ErrorCat(src, " writing symlink target: ", err.Error())
		}
		return nil
	}
	fd, err := os.Open(src)
	if err != nil {
		return utilities.ErrorCat(src, ": opening: ", err.Error())
	}
	defer fd.Close()
	if _, err := io.Copy(writer, fd); err != nil {
		return utilities.ErrorCat(src, ": copying contents: ", err.Error())
	}
	return nil
}
