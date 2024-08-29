package barrow

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	switch b.Compression {
	case "xz":
		zw.RegisterCompressor(XZ, func(w io.Writer) (io.WriteCloser, error) {
			return xz.NewWriter(w)
		})
	case "zstd":
		zw.RegisterCompressor(ZSTD, func(w io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		})
	case "bzip2":
		zw.RegisterCompressor(BZIP2, func(w io.Writer) (io.WriteCloser, error) {
			return bzip2.NewWriter(w, nil)
		})
	case "deflate", "":
	default:
		return zip.Store, fmt.Errorf("unsupported zip compress method '%s'", b.Compression)
	}
	return zip.Deflate, nil
}

func (b *BarrowCtx) addItem2Zip(z *zip.Writer, item *FileItem, method uint16, prefix string) error {
	itemPath := filepath.Join(b.CWD, item.Path)
	si, err := os.Lstat(itemPath)
	if err != nil {
		return err
	}
	var nameInArchive string
	switch {
	case len(item.Rename) != 0:
		nameInArchive = filepath.Join(prefix, item.Destination, item.Rename)
	default:
		nameInArchive = filepath.Join(prefix, item.Destination, filepath.Base(item.Path))
	}

	hdr, err := zip.FileInfoHeader(si)
	if err != nil {
		return err
	}

	if len(item.Permissions) != 0 {
		if m, err := strconv.ParseInt(item.Permissions, 8, 64); err == nil {
			hdr.SetMode(fs.FileMode(m))
		}
	}

	if si.IsDir() {
		hdr.Name = filepath.ToSlash(nameInArchive) + "/"
		hdr.Method = zip.Store
		if _, err = z.CreateHeader(hdr); err != nil {
			return err
		}
		return nil
	}
	hdr.Name = filepath.ToSlash(nameInArchive)
	if isSymlink(si) {
		hdr.SetMode(si.Mode().Perm())
		hdr.Method = Store
		hdr.Modified = si.ModTime()
		w, err := z.CreateHeader(hdr)
		if err != nil {
			return fmt.Errorf("create zip header error: %w", err)
		}
		linkTarget, err := os.Readlink(itemPath)
		if err != nil {
			return fmt.Errorf("add %s to zip error: %w", nameInArchive, err)
		}
		if _, err := w.Write([]byte(filepath.ToSlash(linkTarget))); err != nil {
			return fmt.Errorf("write %s to zip error: %w", linkTarget, err)
		}
		return nil
	}

	hdr.SetMode(mode)
	hdr.Method = method
	hdr.Modified = si.ModTime()
	w, err := z.CreateHeader(hdr)
	if err != nil {
		return fmt.Errorf("create zip header error: %w", err)
	}
	fd, err := os.Open(itemPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	if _, err := io.Copy(w, fd); err != nil {
		return err
	}
	return nil
}

func (b *BarrowCtx) addCrate2Zip(z *zip.Writer, crate *Crate, method uint16, prefix string) error {
	baseName := crate.baseName(b.Target)
	out := filepath.Join(b.Out, crate.Destination, baseName)
	si, err := os.Lstat(out)
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(si)
	if err != nil {
		return err
	}
	nameInArchive := filepath.Join(prefix, crate.Destination, baseName)
	hdr.Name = filepath.ToSlash(nameInArchive)
	hdr.SetMode(0755)
	hdr.Method = method
	hdr.Modified = si.ModTime()
	w, err := z.CreateHeader(hdr)
	if err != nil {
		return fmt.Errorf("create zip header error: %w", err)
	}
	fd, err := os.Open(out)
	if err != nil {
		return err
	}
	defer fd.Close()
	if _, err := io.Copy(w, fd); err != nil {
		return err
	}
	return nil
}

func (b *BarrowCtx) zipInternal(ctx context.Context, p *Package, crates []*Crate, zipPath string, h hash.Hash) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fd, err := os.Create(zipPath)
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
	zipPrefix := strings.TrimSuffix(filepath.Base(zipPath), ".zip")
	_ = z.SetComment(p.Summary)
	for _, item := range p.Include {
		if err := b.addItem2Zip(z, item, method, zipPrefix); err != nil {
			_ = z.Close()
			return err
		}
	}
	for _, crate := range crates {
		if err := b.addCrate2Zip(z, crate, method, zipPrefix); err != nil {
			_ = z.Close()
			return err
		}
	}
	return z.Close()
}

func (b *BarrowCtx) zip(ctx context.Context, p *Package, crates []*Crate) error {
	h := sha256.New()
	zipFileName := fmt.Sprintf("%s-%s-%s-%s.zip", p.Name, p.Version, b.Target, b.Arch)
	var zipPath string
	if filepath.IsAbs(b.Destination) {
		zipPath = filepath.Join(b.Destination, zipFileName)
	} else {
		zipPath = filepath.Join(b.CWD, b.Destination, zipFileName)
	}
	_ = os.MkdirAll(filepath.Dir(zipPath), 0755)
	if err := b.zipInternal(ctx, p, crates, zipPath, h); err != nil {
		fmt.Fprintf(os.Stderr, "zip errpr: %d\n", err)
		_ = os.RemoveAll(zipPath)
		return err
	}
	fmt.Fprintf(os.Stderr, "\x1b[01;36m%s  %s\x1b[0m\n", hex.EncodeToString(h.Sum(nil)), zipFileName)
	return nil
}
