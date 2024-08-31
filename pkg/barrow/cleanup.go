package barrow

import (
	"fmt"
	"os"
	"path/filepath"
)

func (b *BarrowCtx) cleanupCrate(location string) error {
	crate, err := b.LoadCrate(location)
	if err != nil {
		return err
	}
	b.cleanupResources(crate)
	destTo := filepath.Join(b.Out, crate.Destination, crate.Name)
	if err := os.Remove(destTo); err == nil {
		fmt.Fprintf(os.Stderr, "rm: \x1b[33m%s\x1b[0m\n", destTo)
	}
	destToExe := destTo + ".exe"
	if err := os.Remove(destToExe); err == nil {
		fmt.Fprintf(os.Stderr, "rm: \x1b[33m%s\x1b[0m\n", destToExe)
	}
	return nil
}

func (b *BarrowCtx) cleanupItem(item *FileItem, force bool) error {
	saveDir := filepath.Join(b.Out, item.Destination)
	_ = os.MkdirAll(saveDir, 0755)
	source := filepath.Join(b.CWD, item.Path)
	var destTo string
	switch {
	case len(item.Rename) != 0:
		destTo = filepath.Join(saveDir, item.Rename)
	default:
		destTo = filepath.Join(saveDir, filepath.Base(item.Path))
	}
	if si, err := os.Stat(destTo); err == nil {
		o, err := os.Stat(source)
		if err != nil {
			return err
		}
		if si.ModTime().After(o.ModTime()) && !force {
			return nil
		}
	}
	_ = os.Remove(destTo)
	fmt.Fprintf(os.Stderr, "rm: \x1b[33m%s\x1b[0m\n", destTo)
	return nil
}

func (b *BarrowCtx) cleanupPackages() error {
	absDest, err := filepath.Abs(b.Destination)
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(absDest, "*"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, item := range files {
		fmt.Fprintf(os.Stderr, "rm: \x1b[33m%s\x1b[0m\n", item)
		_ = os.Remove(item)
	}
	return nil
}
