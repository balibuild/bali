package barrow

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/rpmpack"
)

// https://docs.redhat.com/zh_hans/documentation/red_hat_enterprise_linux/7/html/rpm_packaging_guide/working-with-spec-files
//

func (b *BarrowCtx) addItem(r *rpmpack.RPM, item *FileItem, prefix string) error {
	fd, err := os.Open(filepath.Join(b.CWD, item.Path))
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
	var saveTo string
	switch {
	case len(item.Rename) != 0:
		saveTo = filepath.Join(prefix, item.Destination, item.Rename)
	default:
		saveTo = filepath.Join(prefix, item.Destination, filepath.Base(item.Path))
	}
	mode := si.Mode().Perm()
	if len(item.Permissions) != 0 {
		if m, err := strconv.ParseInt(item.Permissions, 8, 64); err == nil {
			mode = fs.FileMode(m)
		}
	}
	r.AddFile(rpmpack.RPMFile{
		Name:  saveTo,
		Body:  payload,
		Mode:  uint(mode),
		Group: "root",
		Owner: "root",
		MTime: uint32(si.ModTime().Unix()),
	})
	return nil
}

func (b *BarrowCtx) addCrate(r *rpmpack.RPM, crate *Crate, prefix string) error {
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
	r.AddFile(rpmpack.RPMFile{
		Name:  filepath.Join(prefix, crate.Destination, baseName),
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
)

func (b *BarrowCtx) rpm(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if !rpmSupportedCompressor[b.CompressMethod] {
		return fmt.Errorf("unsupported compressor '%s'", b.CompressMethod)
	}
	packageName := p.PackageName
	if len(packageName) == 0 {
		packageName = p.Name
	}
	// arch := convertArch(b.Arch)
	r, err := rpmpack.NewRPM(rpmpack.RPMMetaData{
		Name:        packageName,
		Summary:     p.Summary,
		Description: p.Description,
		Version:     p.Version,
		Release:     b.Release,
		Arch:        b.Arch,
		Vendor:      p.Vendor,
		URL:         p.URL,
		Packager:    p.Packager,
		Group:       p.Group,
		Licence:     p.License,
		BuildHost:   b.Getenv("BUILD_HOST"),
		Compressor:  b.CompressMethod,
		BuildTime:   time.Now(),
	})
	if err != nil {
		return err
	}
	for _, item := range p.Include {
		if err := b.addItem(r, item, p.Prefix); err != nil {
			return err
		}
	}
	for _, crate := range crates {
		if err := b.addCrate(r, crate, p.Prefix); err != nil {
			return err
		}
	}
	var rpmName string
	if len(r.Release) == 0 {
		rpmName = fmt.Sprintf("%s-%s.%s.rpm", r.Name, r.Version, r.Arch)
	} else {
		rpmName = fmt.Sprintf("%s-%s-%s.%s.rpm", r.Name, r.Version, r.Release, r.Arch)
	}
	var rpmPath string
	if filepath.IsAbs(b.Destination) {
		rpmPath = filepath.Join(b.Destination, rpmName)
	} else {
		rpmPath = filepath.Join(b.CWD, b.Destination, rpmName)
	}
	_ = os.MkdirAll(filepath.Dir(rpmPath), 0755)
	fd, err := os.Create(rpmPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	if err := r.Write(fd); err != nil {
		return err
	}
	return nil
}
