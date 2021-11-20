package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/balibuild/bali/v2/base"
	"github.com/balibuild/bali/v2/builder"
)

// MakeResources todo
func (exe *Executable) MakeResources(wd, syso, binaryName string, be *Executor) error {
	var b builder.Builder
	if len(exe.VersionInfo) != 0 {
		if err := b.ParseJSON(filepath.Join(wd, exe.VersionInfo)); err != nil {
			return base.ErrorCat("pasre versioninfo file: ", err.Error())
		}
	}
	if len(exe.IconPath) != 0 {
		if err := b.AddIcon(filepath.Join(wd, exe.IconPath)); err != nil {
			return base.ErrorCat("load icon: ", err.Error())
		}
	}
	if len(exe.Manifest) != 0 {
		if err := b.AddManifest(filepath.Join(wd, exe.Manifest)); err != nil {
			return base.ErrorCat("load manifest: ", err.Error())
		}
	}
	b.FillVersion(exe.Version, be.bm.Version)
	b.UpdateName(binaryName, be.bm.Name, exe.Description)
	return b.WriteSyso(syso, be.arch)
}

// MakeLinks create links
func (exe *Executable) MakeLinks(destfile string, be *Executor) error {
	for _, l := range exe.Links {
		cl := filepath.Clean(l)
		if len(cl) == 0 || cl == "." {
			continue
		}
		lo := filepath.Join(be.out, be.AddSuffix(cl))
		if base.PathExists(lo) {
			_ = os.RemoveAll(lo)
		}
		if err := os.Symlink(destfile, lo); err != nil {
			return err
		}
		DbgPrint("create symlink %s", lo)
		rel, err := filepath.Rel(filepath.Dir(lo), destfile)
		if err != nil {
			return err
		}
		be.linkmap[cl] = rel
	}
	return nil
}

func (be *Executor) loadExecutable(wd string) (*Executable, error) {
	balisrc := filepath.Join(wd, "balisrc.toml")
	if !base.PathExists(balisrc) {
		fmt.Fprintf(os.Stderr, "%s not found any balisrc.toml\n", wd)
		return nil, os.ErrNotExist
	}
	DbgPrint("%s support toml metadata", wd)
	var exe Executable
	if err := LoadTomlMetadata(balisrc, &exe); err != nil {
		return nil, err
	}
	return &exe, nil
}

// Compile todo
func (be *Executor) Compile(wd string) error {
	exe, err := be.loadExecutable(wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load balisrc from %s error %s\n", wd, err)
		return err
	}
	if len(exe.Name) == 0 {
		exe.Name = filepath.Base(wd)
	} else {
		exe.Name = filepath.Base(exe.Name)
	}
	if exe.Name == "." {
		return base.ErrorCat("bad name: ", exe.Name)
	}
	if be.forceVerion != "" {
		exe.Version = be.forceVerion
	}
	// Update version
	be.UpdateNow(exe.Version)
	binaryName := be.BinaryName(wd, exe.Name)
	var syso string
	if be.target == "windows" {
		syso = builder.MakeSysoPath(wd, be.arch)
		if err := exe.MakeResources(wd, syso, binaryName, be); err != nil {
			return err
		}
	}
	cmd := exec.Command("go", "build", "-o", binaryName)
	for _, s := range exe.GoFlags {
		// Append other args
		cmd.Args = append(cmd.Args, be.ExpandEnv(s))
	}
	cmd.Env = be.environ
	cmd.Dir = wd
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	fmt.Fprintf(os.Stderr, "go compile \x1b[32m%s\x1b[0m version: \x1b[32m%s\x1b[0m\n", exe.Name, exe.Version)
	fmt.Fprintf(os.Stderr, "\x1b[34m%s\x1b[0m\n", cmd.String())
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "compile %s error \x1b[31m%s\x1b[0m\n", exe.Name, err)
		return err
	}
	bindir := filepath.Join(be.out, exe.Destination)
	_ = os.MkdirAll(bindir, 0775)
	binfile := filepath.Join(wd, binaryName)
	destfile := filepath.Join(be.out, exe.Destination, binaryName)
	if err := os.Rename(binfile, destfile); err != nil {
		return err
	}
	be.binaries = append(be.binaries, destfile)
	if len(syso) != 0 {
		_ = os.Remove(syso)
	}
	return exe.MakeLinks(destfile, be)
}
