package barrow

import (
	"context"
	"crypto/sha256"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/apk"
	"github.com/goreleaser/nfpm/v2/arch"
	"github.com/goreleaser/nfpm/v2/deb"
	"github.com/goreleaser/nfpm/v2/files"
)

func (b *BarrowCtx) addItem2Nfpm(info *nfpm.Info, item *FileItem, prefix string) error {
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
		// FIXME: nfpm not write symlink content ?
		info.Overridables.Contents = append(info.Overridables.Contents, &files.Content{
			Type:        files.TypeSymlink,
			Source:      itemPath,
			Destination: nameInArchive,
			FileInfo: &files.ContentFileInfo{
				Mode:  si.Mode(),
				MTime: si.ModTime(),
				Size:  si.Size(),
			},
		})
		return nil
	}
	mode := si.Mode().Perm()
	if len(item.Permissions) != 0 {
		if m, err := strconv.ParseInt(item.Permissions, 8, 64); err == nil {
			mode = fs.FileMode(m)
		}
	}
	info.Overridables.Contents = append(info.Overridables.Contents, &files.Content{
		Source:      itemPath,
		Destination: nameInArchive,
		FileInfo: &files.ContentFileInfo{
			Owner: "root",
			Group: "root",
			Mode:  mode,
			MTime: si.ModTime(),
			Size:  si.Size(),
		},
	})
	return nil
}

func (b *BarrowCtx) addCrate2Nfpm(info *nfpm.Info, crate *Crate, prefix string) error {
	baseName := crate.baseName(b.Target)
	out := filepath.Join(b.Out, crate.Destination, baseName)
	si, err := os.Lstat(out)
	if err != nil {
		return err
	}
	nameInArchive := filepath.Join(prefix, crate.Destination, baseName)
	info.Overridables.Contents = append(info.Overridables.Contents, &files.Content{
		Source:      out,
		Destination: nameInArchive,
		FileInfo: &files.ContentFileInfo{
			Owner: "root",
			Group: "root",
			Mode:  0o755,
			MTime: si.ModTime(),
			Size:  si.Size(),
		},
	})
	return nil
}

func (b *BarrowCtx) deb(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if len(p.Maintainer) == 0 {
		p.Maintainer = "Unset Maintainer <unset@localhost>"
	}
	info := nfpm.WithDefaults(&nfpm.Info{
		Name:        p.Name,
		Platform:    b.Target,
		Arch:        b.Arch,
		Description: p.Description,
		Version:     p.Version,
		Release:     b.Release,
		Maintainer:  p.Maintainer,
		Vendor:      p.Vendor,
		Homepage:    p.Homepage,
		License:     p.License,
	})
	for _, item := range p.Include {
		if err := b.addItem2Nfpm(info, item, p.Prefix); err != nil {
			return err
		}
	}
	for _, crate := range crates {
		if err := b.addCrate2Nfpm(info, crate, p.Prefix); err != nil {
			return err
		}
	}
	debPackageName := deb.Default.ConventionalFileName(info)
	var debPath string
	if filepath.IsAbs(b.Destination) {
		debPath = filepath.Join(b.Destination, debPackageName)
	} else {
		debPath = filepath.Join(b.CWD, b.Destination, debPackageName)
	}
	_ = os.MkdirAll(filepath.Dir(debPath), 0755)
	fd, err := os.Create(debPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	h := sha256.New()
	w := io.MultiWriter(fd, h)
	if err := deb.Default.Package(info, w); err != nil {
		return err
	}
	hashPrint(h, debPackageName)
	return nil
}

func (b *BarrowCtx) apk(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if len(p.Maintainer) == 0 {
		p.Maintainer = "Unset Maintainer <unset@localhost>"
	}
	info := nfpm.WithDefaults(&nfpm.Info{
		Name:        p.Name,
		Platform:    b.Target,
		Arch:        b.Arch,
		Description: p.Description,
		Version:     p.Version,
		Release:     b.Release,
		Maintainer:  p.Maintainer,
		Vendor:      p.Vendor,
		Homepage:    p.Homepage,
		License:     p.License,
	})
	for _, item := range p.Include {
		if err := b.addItem2Nfpm(info, item, p.Prefix); err != nil {
			return err
		}
	}
	for _, crate := range crates {
		if err := b.addCrate2Nfpm(info, crate, p.Prefix); err != nil {
			return err
		}
	}
	apkPackageName := apk.Default.ConventionalFileName(info)
	var apkPath string
	if filepath.IsAbs(b.Destination) {
		apkPath = filepath.Join(b.Destination, apkPackageName)
	} else {
		apkPath = filepath.Join(b.CWD, b.Destination, apkPackageName)
	}
	_ = os.MkdirAll(filepath.Dir(apkPath), 0755)
	fd, err := os.Create(apkPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	h := sha256.New()
	w := io.MultiWriter(fd, h)
	if err := apk.Default.Package(info, w); err != nil {
		return err
	}
	hashPrint(h, apkPackageName)
	return nil
}

func (b *BarrowCtx) archLinux(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if len(p.Maintainer) == 0 {
		p.Maintainer = "Unset Maintainer <unset@localhost>"
	}
	info := nfpm.WithDefaults(&nfpm.Info{
		Name:        p.Name,
		Platform:    b.Target,
		Arch:        b.Arch,
		Description: p.Description,
		Version:     p.Version,
		Release:     b.Release,
		Maintainer:  p.Maintainer,
		Vendor:      p.Vendor,
		Homepage:    p.Homepage,
		License:     p.License,
	})
	for _, item := range p.Include {
		if err := b.addItem2Nfpm(info, item, p.Prefix); err != nil {
			return err
		}
	}
	for _, crate := range crates {
		if err := b.addCrate2Nfpm(info, crate, p.Prefix); err != nil {
			return err
		}
	}
	archLinuxPackageName := arch.Default.ConventionalFileName(info)
	var archLinuxPath string
	if filepath.IsAbs(b.Destination) {
		archLinuxPath = filepath.Join(b.Destination, archLinuxPackageName)
	} else {
		archLinuxPath = filepath.Join(b.CWD, b.Destination, archLinuxPackageName)
	}
	_ = os.MkdirAll(filepath.Dir(archLinuxPath), 0755)
	fd, err := os.Create(archLinuxPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	h := sha256.New()
	w := io.MultiWriter(fd, h)
	if err := arch.Default.Package(info, w); err != nil {
		return err
	}
	hashPrint(h, archLinuxPackageName)
	return nil
}
