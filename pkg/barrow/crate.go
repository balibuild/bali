package barrow

import (
	"fmt"
	"os"
	"path/filepath"
)

type Crate struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description,omitempty"`
	Destination string   `toml:"destination,omitempty"`
	GoFlags     []string `toml:"goflags,omitempty"`
	Version     string   `toml:"version,omitempty"`
	cwd         string   `toml:"-"`
}

func (c *Crate) baseName(target string) string {
	if target == "windows" {
		return c.Name + ".exe"
	}
	return c.Name
}

func (b *BarrowCtx) LoadCrate(location string) (*Crate, error) {
	cwd := filepath.Join(b.CWD, location)
	file := filepath.Join(cwd, "crate.toml")
	var e Crate
	if err := LoadMetadata(file, &e); err != nil {
		return nil, err
	}
	e.cwd = cwd
	if len(e.Name) == 0 {
		e.Name = filepath.Base(cwd)
	}
	if e.Name == "." {
		return nil, fmt.Errorf("unable detect crate name. path '%s'", cwd)
	}
	if len(e.Version) == 0 {
		e.Version = b.Getenv("BUILD_VERSION")
	}
	return &e, nil
}

type WinResCloser func()

func (b *BarrowCtx) MakeResources(e *Crate) (WinResCloser, error) {
	if b.Target != "windows" {
		return nil, nil
	}
	saveTo := filepath.Join(e.cwd, "windows_"+b.Arch+".syso")
	if err := b.makeResources(e, saveTo); err != nil {
		_ = os.RemoveAll(saveTo)
		return nil, err
	}
	return func() {
		// remove
		_ = os.RemoveAll(saveTo)
	}, nil
}
