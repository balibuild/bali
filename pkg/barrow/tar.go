package barrow

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

func (b *BarrowCtx) addItem2Tar(z *tar.Writer, item *FileItem, prefix string) error {
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
	var linkTarget string
	if isSymlink(si) {
		if linkTarget, err = os.Readlink(itemPath); err != nil {
			return err
		}
	}
	hdr, err := tar.FileInfoHeader(si, linkTarget)
	if err != nil {
		return err
	}
	if len(item.Permissions) != 0 {
		if m, err := strconv.ParseInt(item.Permissions, 8, 64); err == nil {
			hdr.Mode = m
		}
	}
	hdr.Name = AsExplicitRelativePath(nameInArchive)
	if err = z.WriteHeader(hdr); err != nil {
		return fmt.Errorf("write tar header error: %w", err)
	}
	if si.IsDir() {
		return nil
	}
	if hdr.Typeflag != tar.TypeReg {
		return nil
	}
	fd, err := os.Open(itemPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	if _, err := io.Copy(z, fd); err != nil {
		return err
	}
	return nil
}

func (b *BarrowCtx) addCrate2Tar(z *tar.Writer, crate *Crate, prefix string) error {
	baseName := crate.baseName(b.Target)
	out := filepath.Join(b.Out, crate.Destination, baseName)
	si, err := os.Lstat(out)
	if err != nil {
		return err
	}
	nameInArchive := filepath.Join(prefix, crate.Destination, baseName)
	hdr, err := tar.FileInfoHeader(si, "")
	if err != nil {
		return err
	}
	hdr.Name = AsExplicitRelativePath(nameInArchive)
	hdr.Mode = 0755
	if err = z.WriteHeader(hdr); err != nil {
		return fmt.Errorf("write tar header error: %w", err)
	}
	fd, err := os.Open(out)
	if err != nil {
		return err
	}
	defer fd.Close()
	if _, err := io.Copy(z, fd); err != nil {
		return err
	}
	return nil
}

func (b *BarrowCtx) tarInternal(p *Package, crates []*Crate, prefix string, w io.Writer) error {
	z := tar.NewWriter(w)
	for _, item := range p.Include {
		if err := b.addItem2Tar(z, item, prefix); err != nil {
			_ = z.Close()
			return err
		}
	}
	for _, crate := range crates {
		if err := b.addCrate2Tar(z, crate, prefix); err != nil {
			_ = z.Close()
			return err
		}
	}
	return z.Close()
}

//go:embed resources/template.sh
var resources embed.FS

func (b *BarrowCtx) sh(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	newCompressor, _, err := tarCompressor(b.Compression)
	if err != nil {
		return err
	}
	tarFileName := fmt.Sprintf("%s-%s-%s-%s.sh", p.Name, p.Version, b.Target, b.Arch)
	var tarPath string
	if filepath.IsAbs(b.Destination) {
		tarPath = filepath.Join(b.Destination, tarFileName)
	} else {
		tarPath = filepath.Join(b.CWD, b.Destination, tarFileName)
	}
	_ = os.MkdirAll(filepath.Dir(tarPath), 0755)
	fd, err := os.OpenFile(tarPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	h := sha256.New()
	w := io.MultiWriter(fd, h)
	rfd, err := resources.Open("resources/template.sh")
	if err != nil {
		return err
	}
	if _, err := fd.ReadFrom(rfd); err != nil {
		_ = fd.Close()
		return err
	}
	cw, err := newCompressor(w)
	if err != nil {
		_ = fd.Close()
		return err
	}
	if err := b.tarInternal(p, crates, "", cw); err != nil {
		fmt.Fprintf(os.Stderr, "zip errpr: %d\n", err)
		_ = cw.Close()
		_ = fd.Close()
		_ = os.RemoveAll(tarPath)
		return err
	}
	if err := cw.Close(); err != nil {
		_ = fd.Close()
		_ = os.RemoveAll(tarPath)
		return err
	}
	fmt.Fprintf(os.Stderr, "\x1b[01;36m%s  %s\x1b[0m\n", hex.EncodeToString(h.Sum(nil)), tarFileName)
	return nil
}

type FnCompressor func(w io.Writer) (io.WriteCloser, error)

func tarCompressor(method string) (FnCompressor, string, error) {
	switch method {
	case "zstd":
		return func(w io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		}, ".tar.zst", nil
	case "xz":
		return func(w io.Writer) (io.WriteCloser, error) {
			return xz.NewWriter(w)
		}, ".tar.xz", nil
	case "bzip2":
		return func(w io.Writer) (io.WriteCloser, error) {
			return bzip2.NewWriter(w, nil)
		}, ".tar.bz2", nil
	case "brotli":
		return func(w io.Writer) (io.WriteCloser, error) {
			return brotli.NewWriter(w), nil
		}, ".tar.br", nil
	case "", "gzip":
		return func(w io.Writer) (io.WriteCloser, error) {
			return gzip.NewWriter(w), nil
		}, ".tar.gz", nil
	case "none":
		return func(w io.Writer) (io.WriteCloser, error) {
			return &nopCloser{Writer: w}, nil
		}, ".tar", nil
	default:
		return nil, "", fmt.Errorf("unsupported tar compress method '%s'", method)
	}
}

func (b *BarrowCtx) tar(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	newCompressor, suffix, err := tarCompressor(b.Compression)
	if err != nil {
		return err
	}
	tarPrefix := fmt.Sprintf("%s-%s-%s-%s", p.Name, p.Version, b.Target, b.Arch)
	tarFileName := tarPrefix + suffix
	var tarPath string
	if filepath.IsAbs(b.Destination) {
		tarPath = filepath.Join(b.Destination, tarFileName)
	} else {
		tarPath = filepath.Join(b.CWD, b.Destination, tarFileName)
	}
	_ = os.MkdirAll(filepath.Dir(tarPath), 0755)
	fd, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	h := sha256.New()
	cw, err := newCompressor(io.MultiWriter(fd, h))
	if err != nil {
		_ = fd.Close()
		return err
	}
	if err := b.tarInternal(p, crates, tarPrefix, cw); err != nil {
		fmt.Fprintf(os.Stderr, "zip errpr: %d\n", err)
		_ = cw.Close()
		_ = fd.Close()
		_ = os.RemoveAll(tarPath)
		return err
	}
	if err := cw.Close(); err != nil {
		_ = fd.Close()
		_ = os.RemoveAll(tarPath)
		return err
	}
	fmt.Fprintf(os.Stderr, "\x1b[01;36m%s  %s\x1b[0m\n", hex.EncodeToString(h.Sum(nil)), tarFileName)
	return nil
}
