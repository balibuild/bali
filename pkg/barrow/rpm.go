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

// var (
// 	archList = map[string]string{
// 		"arm64":   "aarch64",
// 		"amd64":   "x86_64",
// 		"riscv64": "riscv64",
// 		"loong64": "loong64",
// 	}
// )

// func convertArch(arch string) string {
// 	if a, ok := archList[arch]; ok {
// 		return a
// 	}
// 	return arch
// }

func (b *BarrowCtx) rpm(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
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
	switch {
	case len(r.Release) == 0:
		rpmName = fmt.Sprintf("%s-%s.%s.rpm", r.Name, r.Version, r.Arch)
	default:
		rpmName = fmt.Sprintf("%s-%s-%s.%s.rpm", r.Name, r.Version, r.Release, r.Arch)
	}
	fd, err := os.Create(filepath.Join(b.CWD, rpmName))
	if err != nil {
		return err
	}
	defer fd.Close()
	if err := r.Write(fd); err != nil {
		return err
	}
	return nil
}
