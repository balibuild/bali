package pack

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/balibuild/bali/utilities"
)

// Zip
const (
	ZipISVTX = 0x200
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

// AddTargetLink create zip symlink
func (zp *ZipPacker) AddTargetLink(nameInArchive, linkName string) error {
	var hdr zip.FileHeader
	hdr.SetModTime(time.Now())
	hdr.SetMode(0777) // symlink
	hdr.Name = filepath.ToSlash(nameInArchive)
	writer, err := zp.zw.CreateHeader(&hdr)
	if err != nil {
		return utilities.ErrorCat(linkName, ": making header:", err.Error())
	}
	if _, err := writer.Write([]byte(filepath.ToSlash(linkName))); err != nil {
		return utilities.ErrorCat(linkName, " writing symlink target: ", err.Error())
	}
	return nil
}

// AddFileEx todo
func (zp *ZipPacker) AddFileEx(src, nameInArchive string, exerights bool) error {
	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(st)
	if err != nil {
		return utilities.ErrorCat(src, ": getting header: ", err.Error())
	}
	if exerights {
		header.SetMode(header.Mode() | 0755)
	}
	if st.IsDir() {
		// Windows support '/'
		header.Name = utilities.StrCat(filepath.ToSlash(nameInArchive), "/")
		header.Method = zip.Store
	} else {
		header.Name = filepath.ToSlash(nameInArchive)
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

// AddFile file to zip packer
func (zp *ZipPacker) AddFile(src, nameInArchive string) error {
	return zp.AddFileEx(src, nameInArchive, false)
}
