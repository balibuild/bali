package pack

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/balibuild/bali/utilities"
	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// Zip
const (
	ZipISVTX = 0x200
)

// Compression methods.
const (
	Store   uint16 = 0  // no compression
	Deflate uint16 = 8  // DEFLATE compressed
	BZIP2   uint16 = 12 // bzip2
	LZMA    uint16 = 14 //LZMA
	ZSTD    uint16 = 93 //see https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT.
	XZ      uint16 = 95
)

// ZipPacker todo
type ZipPacker struct {
	zw         *zip.Writer
	FileMethod uint16 // zip filemethod
}

// NewZipPacker todo
func NewZipPacker(w io.Writer) *ZipPacker {
	return &ZipPacker{zw: zip.NewWriter(w), FileMethod: Deflate}
}

// NewZipPackerEx todo
func NewZipPackerEx(w io.Writer, method uint16) *ZipPacker {
	zp := NewZipPacker(w)
	switch method {
	case BZIP2:
		zp.zw.RegisterCompressor(BZIP2, func(out io.Writer) (io.WriteCloser, error) {
			return bzip2.NewWriter(out, nil)
		})
		zp.FileMethod = BZIP2
	case ZSTD:
		zp.zw.RegisterCompressor(ZSTD, func(out io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedFastest))
		})
		zp.FileMethod = ZSTD
	case XZ:
		zp.zw.RegisterCompressor(XZ, func(out io.Writer) (io.WriteCloser, error) {
			return xz.NewWriter(out)
		})
		zp.FileMethod = XZ
	}
	return zp
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
	hdr.SetMode(0755 | os.ModeSymlink) // symlink
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
		header.Method = zp.FileMethod
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
