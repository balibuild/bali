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
	Alias       []string `toml:"alias,omitempty"` // with out suffix
	cwd         string   `toml:"-"`
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
		_ = os.Remove(saveTo)
		return nil, err
	}
	return func() {
		_ = os.Remove(saveTo)
	}, nil
}

func (b *BarrowCtx) cleanupResources(e *Crate) {
	files, err := filepath.Glob(filepath.Join(e.cwd, "*.syso"))
	if err != nil {
		return
	}
	for _, item := range files {
		fmt.Fprintf(os.Stderr, "rm: \x1b[33m%s\x1b[0m\n", item)
		_ = os.Remove(item)
	}
}
