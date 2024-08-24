package pack

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/balibuild/bali/v3/base"
	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/compress/zstd"
)

// Zip
const (
	ZipISVTX = 0x200
)

// Compression methods.
const (
	Store   uint16 = 0   // no compression
	Deflate uint16 = 8   // DEFLATE compressed
	BZIP2   uint16 = 12  // bzip2
	LZMA    uint16 = 14  // LZMA
	ZSTD    uint16 = 93  // see https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT.
	XZ      uint16 = 95  // XZ
	BROTLI  uint16 = 121 // private
)

// ZipBuilder todo
type ZipBuilder struct {
	zw     *zip.Writer
	Method uint16 // zip filemethod
}

// NewZipBuilder todo
func NewZipBuilder(w io.Writer) *ZipBuilder {
	return &ZipBuilder{zw: zip.NewWriter(w), Method: Deflate}
}

// NewZipBuilderEx with compress method
func NewZipBuilderEx(w io.Writer, method uint16) *ZipBuilder {
	b := NewZipBuilder(w)
	switch method {
	case BZIP2:
		b.zw.RegisterCompressor(BZIP2, func(out io.Writer) (io.WriteCloser, error) {
			return bzip2.NewWriter(out, nil)
		})
		b.Method = BZIP2
	case ZSTD:
		b.zw.RegisterCompressor(ZSTD, func(out io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
			//return zstd.ZipCompressor()(out)
		})
		b.Method = ZSTD
	case BROTLI:
		b.zw.RegisterCompressor(BROTLI, func(out io.Writer) (io.WriteCloser, error) {
			return brotli.NewWriter(out), nil
		})
		b.Method = BROTLI
	default:
	}
	return b
}

// Close todo
func (b *ZipBuilder) Close() error {
	if b.zw == nil {
		return nil
	}
	return b.zw.Close()
}

func (zp *ZipBuilder) SetComment(comment string) error {
	return zp.zw.SetComment(comment)
}

// AddTargetLink create zip symlink
func (zp *ZipBuilder) AddTargetLink(nameInArchive, linkName string) error {
	var hdr zip.FileHeader
	hdr.Modified = time.Now()
	hdr.SetMode(0755 | os.ModeSymlink) // symlink
	hdr.Name = filepath.ToSlash(nameInArchive)
	writer, err := zp.zw.CreateHeader(&hdr)
	if err != nil {
		return base.ErrorCat(linkName, ": making header:", err.Error())
	}
	if _, err := writer.Write([]byte(filepath.ToSlash(linkName))); err != nil {
		return base.ErrorCat(linkName, " writing symlink target: ", err.Error())
	}
	return nil
}

// AddFileEx todo
func (zp *ZipBuilder) AddFileEx(src, nameInArchive string, exerights bool) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(fi)
	if err != nil {
		return base.ErrorCat(src, ": getting header: ", err.Error())
	}
	if fi.IsDir() {
		hdr.Name = base.StrCat(filepath.ToSlash(nameInArchive), "/")
		hdr.Method = zip.Store
		if _, err = zp.zw.CreateHeader(hdr); err != nil {
			return base.ErrorCat(nameInArchive, ": making header:", err.Error())
		}
		return nil
	}
	if exerights {
		hdr.SetMode(hdr.Mode() | 0755)
	}
	hdr.Name = filepath.ToSlash(nameInArchive)
	hdr.Method = zp.Method
	writer, err := zp.zw.CreateHeader(hdr)
	if err != nil {
		return base.ErrorCat(nameInArchive, ": making header:", err.Error())
	}
	if isSymlink(fi) {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return base.ErrorCat(src, ": readlink: ", err.Error())
		}
		if _, err := writer.Write([]byte(filepath.ToSlash(linkTarget))); err != nil {
			return base.ErrorCat(src, " writing symlink target: ", err.Error())
		}
		return nil
	}
	fd, err := os.Open(src)
	if err != nil {
		return base.ErrorCat(src, ": opening: ", err.Error())
	}
	defer fd.Close()
	if _, err := io.Copy(writer, fd); err != nil {
		return base.ErrorCat(src, ": copying contents: ", err.Error())
	}
	return nil
}

// AddFile file to zip packer
func (zp *ZipBuilder) AddFile(src, nameInArchive string) error {
	return zp.AddFileEx(src, nameInArchive, false)
}
