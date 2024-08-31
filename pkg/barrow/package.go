package barrow

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

type FileItem struct {
	Path        string `toml:"path"`
	Destination string `toml:"destination"`
	Rename      string `toml:"rename,omitempty"`      // when rename is no empty: rename file to name
	Permissions string `toml:"permissions,omitempty"` // 0755 0644
}

type Package struct {
	Name        string      `toml:"name"`
	PackageName string      `toml:"package-name,omitempty"`
	Summary     string      `toml:"summary,omitempty"`     // Is a short description of the software
	Description string      `toml:"description,omitempty"` // description is a longer piece of software information than Summary, consisting of one or more paragraphs
	Version     string      `toml:"version,omitempty"`
	Authors     []string    `toml:"authors,omitempty"`
	Vendor      string      `toml:"vendor,omitempty"`
	Maintainer  string      `toml:"maintainer,omitempty"`
	Homepage    string      `toml:"homepage,omitempty"`
	Packager    string      `toml:"packager,omitempty"` // BALI_RPM_PACKAGER
	Group       string      `toml:"group,omitempty"`
	License     string      `toml:"license,omitempty"`
	LicenseFile string      `toml:"license-file,omitempty"`
	Prefix      string      `toml:"prefix,omitempty"` // install prefix: rpm required
	Crates      []string    `toml:"crates,omitempty"`
	Include     []*FileItem `toml:"include,omitempty"`
}

func LoadMetadata(file string, v any) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	if err := toml.NewDecoder(fd).Decode(v); err != nil {
		return err
	}
	return nil
}

func (b *BarrowCtx) LoadPackage(cwd string) (*Package, error) {
	file := filepath.Join(cwd, "bali.toml")
	var p Package
	if err := LoadMetadata(file, &p); err != nil {
		return nil, err
	}
	if packageName, ok := os.LookupEnv("PACKAGE_NAME"); ok {
		p.PackageName = packageName // overwrite
	}
	// OS BUILD_VERSION
	if version, ok := os.LookupEnv("BUILD_VERSION"); ok {
		p.Version = version
	}
	return &p, nil
}

func copyTo(src, dest string, newPerm string) error {
	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("copyTo: non-regular source file %s (%q)", st.Name(), st.Mode().String())
	}
	dst, err := os.Stat(dest)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !dst.Mode().IsRegular() {
			return fmt.Errorf("copyTo: non-regular destination file %s (%q)", dst.Name(), dst.Mode().String())
		}
		if os.SameFile(st, dst) {
			return nil
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	perm := st.Mode().Perm()
	if len(newPerm) != 0 {
		if m, err := strconv.ParseInt(newPerm, 8, 64); err == nil {
			perm = fs.FileMode(m)
		}
	}
	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func (b *BarrowCtx) apply(item *FileItem) error {
	saveDir := filepath.Join(b.Out, item.Destination)
	_ = os.MkdirAll(saveDir, 0755)
	source := filepath.Join(b.CWD, item.Path)
	var saveTo string
	switch {
	case len(item.Rename) != 0:
		saveTo = filepath.Join(saveDir, item.Rename)
	default:
		saveTo = filepath.Join(saveDir, filepath.Base(item.Path))
	}
	if si, err := os.Stat(saveTo); err == nil {
		o, err := os.Stat(source)
		if err != nil {
			return err
		}
		if si.ModTime().After(o.ModTime()) {
			return nil
		}
	}
	if err := copyTo(source, saveTo, item.Permissions); err != nil {
		fmt.Fprintf(os.Stderr, "install %s error: %v\n", item.Path, err)
		return err
	}
	if len(item.Rename) != 0 {
		stage("install", "\x1b[38;02;39;199;173m%s\x1b[0m --> \x1b[38;02;39;199;173m%s\x1b[0m done", item.Path, filepath.Join(item.Destination, item.Rename))
		return nil
	}
	stage("install", "\x1b[38;02;39;199;173m%s\x1b[0m --> \x1b[38;02;39;199;173m%s\x1b[0m done", item.Path, filepath.Base(item.Path))
	return nil
}
