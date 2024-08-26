package barrow

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

const (
	Store   uint16 = 0   // no compression
	Deflate uint16 = 8   // DEFLATE compressed
	BZIP2   uint16 = 12  // bzip2
	LZMA    uint16 = 14  // LZMA
	ZSTD    uint16 = 93  // see https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT.
	XZ      uint16 = 95  // XZ
	BROTLI  uint16 = 121 // private
)

func (b *BarrowCtx) registerCompressor(zw *zip.Writer) (uint16, error) {
	switch b.CompressMethod {
	case "xz":
		zw.RegisterCompressor(XZ, func(w io.Writer) (io.WriteCloser, error) {
			return xz.NewWriter(w)
		})
	case "zstd":
		zw.RegisterCompressor(ZSTD, func(w io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		})
	case "bzip":
		zw.RegisterCompressor(BZIP2, func(w io.Writer) (io.WriteCloser, error) {
			return bzip2.NewWriter(w, nil)
		})
	case "deflate", "":
	default:
		return zip.Store, fmt.Errorf("unsupported zip compress method '%s'", b.CompressMethod)
	}
	return zip.Deflate, nil
}

func (b *BarrowCtx) zipInternal(ctx context.Context, p *Package, crates []*Crate, dist string, h hash.Hash) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fd, err := os.Create(dist)
	if err != nil {
		return err
	}
	defer fd.Close()
	w := io.MultiWriter(fd, h)
	z := zip.NewWriter(w) // TODO
	method, err := b.registerCompressor(z)
	if err != nil {
		return err
	}
	_ = z.SetComment(p.Summary)
	fmt.Fprintf(os.Stderr, "method: %d\n", method)
	for _, crate := range crates {
		fmt.Fprintf(os.Stderr, "%s\n", crate.Name)
	}
	return nil
}

func (b *BarrowCtx) zip(ctx context.Context, p *Package, crates []*Crate) error {
	h := sha256.New()
	zipName := fmt.Sprintf("%s-%s-%s-%s.zip", p.Name, p.Version, b.Target, b.Arch)
	var zipPath string
	if filepath.IsAbs(b.Destination) {
		zipPath = filepath.Join(b.Destination, zipName)
	} else {
		zipPath = filepath.Join(b.CWD, b.Destination, zipName)
	}
	_ = os.MkdirAll(filepath.Dir(zipPath), 0755)
	if err := b.zipInternal(ctx, p, crates, zipPath, h); err != nil {
		fmt.Fprintf(os.Stderr, "zip errpr: %d\n", err)
		_ = os.RemoveAll(zipPath)
		return err
	}
	fmt.Fprintf(os.Stderr, "\x1b[34m%s  %s\x1b[0m\n", hex.EncodeToString(h.Sum(nil)), zipName)
	return nil
}
