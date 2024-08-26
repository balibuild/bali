package barrow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/rpmpack"
)

// https://docs.redhat.com/zh_hans/documentation/red_hat_enterprise_linux/7/html/rpm_packaging_guide/working-with-spec-files

const (
	// Symbolic link
	tagLink = 0o120000
)

func (b *BarrowCtx) addItem2RPM(r *rpmpack.RPM, item *FileItem, prefix string) error {
	itemPath := filepath.Join(b.CWD, item.Path)
	var nameInArchive string
	switch {
	case len(item.Rename) != 0:
		nameInArchive = filepath.Join(prefix, item.Destination, item.Rename)
	default:
		nameInArchive = filepath.Join(prefix, item.Destination, filepath.Base(item.Path))
	}
	si, err := os.Lstat(itemPath)
	if err != nil {
		return err
	}
	if isSymlink(si) {
		linkTarget, err := os.Readlink(itemPath)
		if err != nil {
			return fmt.Errorf("add %s to zip error: %w", nameInArchive, err)
		}
		r.AddFile(rpmpack.RPMFile{
			Name:  filepath.ToSlash(nameInArchive),
			Body:  []byte(filepath.ToSlash(linkTarget)),
			Mode:  tagLink,
			Group: "root",
			Owner: "root",
			MTime: uint32(si.ModTime().Unix()),
		})
		return nil
	}
	mode := si.Mode().Perm()
	if len(item.Permissions) != 0 {
		if m, err := strconv.ParseInt(item.Permissions, 8, 64); err == nil {
			mode = fs.FileMode(m)
		}
	}
	fd, err := os.Open(itemPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	payload, err := io.ReadAll(fd)
	if err != nil {
		return err
	}
	r.AddFile(rpmpack.RPMFile{
		Name:  filepath.ToSlash(nameInArchive),
		Body:  payload,
		Mode:  uint(mode),
		Group: "root",
		Owner: "root",
		MTime: uint32(si.ModTime().Unix()),
	})
	return nil
}

func (b *BarrowCtx) addCrate2RPM(r *rpmpack.RPM, crate *Crate, prefix string) error {
	baseName := crate.baseName(b.Target)
	out := filepath.Join(b.Out, crate.Destination, baseName)
	fd, err := os.Open(out)
	if err != nil {
		return err
	}
	defer fd.Close()
	si, err := fd.Stat()
	if err != nil {
		return err
	}
	payload, err := io.ReadAll(fd)
	if err != nil {
		return err
	}
	nameInArchive := filepath.Join(prefix, crate.Destination, baseName)
	r.AddFile(rpmpack.RPMFile{
		Name:  filepath.ToSlash(nameInArchive),
		Body:  payload,
		Mode:  0755,
		Group: "root",
		Owner: "root",
		MTime: uint32(si.ModTime().Unix()),
	})
	return nil
}

var (
	rpmSupportedCompressor = map[string]bool{
		"":     true,
		"gzip": true,
		"zstd": true,
		"lzma": true,
		"xz":   true,
	}
	// https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch01s03.html
	// nolint: gochecknoglobals
	rpmArchList = map[string]string{
		"all":      "noarch",
		"amd64":    "x86_64",
		"386":      "i386",
		"arm64":    "aarch64",
		"arm5":     "armv5tel",
		"arm6":     "armv6hl",
		"arm7":     "armv7hl",
		"mips64le": "mips64el",
		"mipsle":   "mipsel",
		"mips":     "mips",
		// TODO: other arches
	}
)

func rpmArchName(arch string) string {
	if a, ok := rpmArchList[arch]; ok {
		return a
	}
	return arch
}

func (b *BarrowCtx) rpm(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if !rpmSupportedCompressor[b.Compression] {
		return fmt.Errorf("unsupported compressor '%s'", b.Compression)
	}
	r, err := rpmpack.NewRPM(rpmpack.RPMMetaData{
		Name:        nonEmpty(p.PackageName, p.Name),
		Summary:     nonEmpty(p.Summary, strings.Split(p.Description, "\n")[0]),
		Description: p.Description,
		Version:     p.Version,
		Release:     nonEmpty(b.Release, "1"),
		Arch:        rpmArchName(b.Arch),
		Vendor:      p.Vendor,
		URL:         p.Homepage,
		Packager:    p.Packager,
		Group:       p.Group,
		Licence:     p.License,
		BuildHost:   b.Getenv("BUILD_HOST"),
		Compressor:  b.Compression,
		BuildTime:   time.Now(),
	})
	if err != nil {
		return err
	}
	for _, item := range p.Include {
		if err := b.addItem2RPM(r, item, p.Prefix); err != nil {
			return err
		}
	}
	for _, crate := range crates {
		if err := b.addCrate2RPM(r, crate, p.Prefix); err != nil {
			return err
		}
	}
	rpmPackageName := fmt.Sprintf("%s-%s-%s.%s.rpm", r.Name, r.Version, r.Release, r.Arch)
	var rpmPath string
	if filepath.IsAbs(b.Destination) {
		rpmPath = filepath.Join(b.Destination, rpmPackageName)
	} else {
		rpmPath = filepath.Join(b.CWD, b.Destination, rpmPackageName)
	}
	_ = os.MkdirAll(filepath.Dir(rpmPath), 0755)
	fd, err := os.Create(rpmPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	h := sha256.New()
	w := io.MultiWriter(fd, h)
	if err := r.Write(w); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\x1b[01;36m%s  %s\x1b[0m\n", hex.EncodeToString(h.Sum(nil)), rpmPackageName)
	return nil
}
