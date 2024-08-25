package barrow

import (
	"context"
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

func (b *BarrowCtx) LoadCrate(location string) (*Crate, error) {
	cwd := filepath.Join(b.CWD, location)
	file := filepath.Join(cwd, "crate.toml")
	var e Crate
	if err := LoadMetadata(file, &e); err != nil {
		return nil, err
	}
	e.cwd = cwd
	return &e, nil
}

func (b *BarrowCtx) MakeWinRes(ctx context.Context, e *Crate) error {
	if b.Target != "windows" {
		return nil
	}
	return nil
}
