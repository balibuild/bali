package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/balibuild/bali/v2/base"
	"github.com/balibuild/bali/v2/pack"
)

// Executor

// Executor todo
type Executor struct {
	de             *base.Derivator
	target         string // os
	arch           string
	out            string
	workdir        string
	destination    string
	makezip        bool
	zipmethod      uint16
	makepack       bool
	norename       bool
	cleanup        bool
	environ        []string // initialize environment
	binaries       []string
	linkmap        map[string]string
	bm             Project
	suffix         string
	forceVerion    string
	withoutVersion bool
}

func resolveBuildID(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = cwd
	if out, err := cmd.CombinedOutput(); err == nil {
		commitid := strings.TrimSpace(string(out))
		DbgPrint("BUILD_COMMIT: '%s'", commitid)
		return commitid
	}
	return "None"
}

func resolveReference(cwd string) string {
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = cwd
	if out, err := cmd.CombinedOutput(); err == nil {
		ref := strings.TrimSpace(string(out))
		DbgPrint("BUILD_REFNAME: '%s'", ref)
		return ref
	}
	// Github support
	if refName := os.Getenv("GITHUB_REF"); len(refName) != 0 {
		return refName
	}
	if refName := os.Getenv("GIT_BUILD_REF"); len(refName) != 0 {
		return refName
	}
	return "None"
}

func resolveGoVersion() string {
	cmd := exec.Command("go", "version")
	if out, err := cmd.CombinedOutput(); err == nil {
		goversion := strings.TrimPrefix(strings.TrimSpace(string(out)), "go version go")
		DbgPrint("Go Version: '%s'", goversion)
		return goversion
	}
	return "None"
}

func resolveDistSupport(target, arch string) bool {
	cmd := exec.Command("go", "tool", "dist", "list")
	if out, err := cmd.CombinedOutput(); err == nil {
		str := base.StrCat(target, "/", arch)
		if strings.Contains(string(out), str) {
			return true
		}
	}
	return false
}

func (be *Executor) initializeProject() error {
	balitoml := filepath.Join(be.workdir, "bali.toml")
	if !base.PathExists(balitoml) {
		fmt.Fprintf(os.Stderr, "%s not found 'bali.toml'\n", be.workdir)
		return os.ErrNotExist
	}
	DbgPrint("%s support toml metadata", be.workdir)
	return LoadTomlMetadata(balitoml, &be.bm)
}

// Initialize todo
func (be *Executor) Initialize() error {
	if err := be.initializeProject(); err != nil {
		return err
	}
	if len(be.forceVerion) != 0 {
		be.bm.Version = be.forceVerion
	}
	if len(be.bm.Version) == 0 {
		be.bm.Version = "0.0.1"
	}
	be.de = base.NewDerivator()
	// Respect environmental variable settings
	if len(be.target) == 0 {
		if be.target = os.Getenv("GOOS"); len(be.target) == 0 {
			be.target = runtime.GOOS
		}
	}
	if len(be.arch) == 0 {
		if be.arch = os.Getenv("GOARCH"); len(be.arch) == 0 {
			be.arch = runtime.GOARCH
		}
	}
	if !resolveDistSupport(be.target, be.arch) {
		return base.ErrorCat("unsupported GOOS/GOARCH pair ", be.target, "/", be.arch)
	}
	refname := resolveReference(be.workdir)
	_ = be.de.Append("BUILD_REFNAME", refname)
	_ = be.de.Append("BUILD_COMMIT", resolveBuildID(be.workdir))
	if strings.HasPrefix(refname, "refs/heads/") {
		_ = be.de.Append("BUILD_BRANCH", strings.TrimPrefix(refname, "refs/heads/"))
	} else {
		_ = be.de.Append("BUILD_BRANCH", "")
	}
	_ = be.de.Append("BUILD_GOVERSION", resolveGoVersion())
	_ = be.de.Append("BUILD_VERSION", "0.0.1")
	t := time.Now()
	_ = be.de.Append("BUILD_TIME", t.Format(time.RFC3339))
	_ = be.de.Append("BUILD_YEAR", strconv.Itoa(t.Year()))
	osenv := os.Environ()
	be.environ = make([]string, 0, len(osenv)+3)
	for _, e := range osenv {
		// remove old env
		if !strings.HasPrefix(e, "GOOS=") && !strings.HasPrefix(e, "GOARCH=") {
			be.environ = append(be.environ, e)
		}
	}
	be.environ = append(be.environ, base.StrCat("GOOS=", be.target))
	be.environ = append(be.environ, base.StrCat("GOARCH=", be.arch))
	be.linkmap = make(map[string]string)
	if be.target == "windows" {
		be.suffix = ".exe"
	}
	return nil
}

// UpdateNow todo
func (be *Executor) UpdateNow(version string) {
	_ = be.de.Append("BUILD_VERSION", version)
	t := time.Now()
	_ = be.de.Append("BUILD_TIME", t.Format(time.RFC3339))
}

// ExpandEnv todo
func (be *Executor) ExpandEnv(s string) string {
	return be.de.ExpandEnv(s)
}

// TargetName todo
func (be *Executor) TargetName(name string) string {
	if be.norename {
		return name
	}
	return base.StrCat(name, ".new")
}

// FileName todo
func (be *Executor) FileName(file *File) string {
	if be.norename || file.NoRename {
		return file.Base()
	}
	return base.StrCat(file.Base(), ".template")
}

// AddSuffix todo
func (be *Executor) AddSuffix(name string) string {
	if len(be.suffix) == 0 {
		return name
	}
	if strings.HasSuffix(name, be.suffix) {
		return name
	}
	return base.StrCat(name, be.suffix)
}

// BinaryName todo
func (be *Executor) BinaryName(dir, name string) string {
	if len(name) == 0 {
		return base.StrCat(filepath.Base(dir), be.suffix)
	}
	return base.StrCat(name, be.suffix)
}

// PathInArchive todo
func (be *Executor) PathInArchive(destination string) string {
	if be.makepack {
		return destination
	}
	if be.withoutVersion {
		return filepath.Join(base.StrCat(be.bm.Name, "-", be.target, "-", be.arch), destination)
	}
	return filepath.Join(base.StrCat(be.bm.Name, "-", be.target, "-", be.arch, "-", be.bm.Version), destination)
}

// Build todo
func (be *Executor) Build() error {
	if err := be.bm.FileConfigure(be.workdir, be.out); err != nil {
		return err
	}
	for _, d := range be.bm.Dirs {
		wd := filepath.Join(be.workdir, d)
		if err := be.Compile(wd); err != nil {
			fmt.Fprintf(os.Stderr, "bali compile: \x1b[31m%s\x1b[0m\n", err)
			return err
		}
	}
	return nil
}

// Compress todo
func (be *Executor) Compress() error {
	if !base.PathDirExists(be.destination) {
		if err := os.MkdirAll(be.destination, 0755); err != nil {
			return err
		}
	}
	var outfile string
	var err error
	var fd *os.File
	var mw io.Writer
	var pk pack.Packer
	h := sha256.New()
	var outname string
	if be.target == "windows" {
		if be.withoutVersion {
			outname = base.StrCat(be.bm.Name, "-", be.target, "-", be.arch, ".zip")
		} else {
			outname = base.StrCat(be.bm.Name, "-", be.target, "-", be.arch, "-", be.bm.Version, ".zip")
		}
		outfile = filepath.Join(be.destination, outname)
		if fd, err = os.Create(outfile); err != nil {
			return err
		}
		mw = io.MultiWriter(fd, h)
		zpk := pack.NewZipPackerEx(mw, be.zipmethod)
		if len(be.bm.Destination) != 0 {
			zpk.SetComment(be.bm.Destination)
		}
		pk = zpk
	} else {
		if be.withoutVersion {
			outname = base.StrCat(be.bm.Name, "-", be.target, "-", be.arch, ".tar.gz")
		} else {
			outname = base.StrCat(be.bm.Name, "-", be.target, "-", be.arch, "-", be.bm.Version, ".tar.gz")
		}

		outfile = filepath.Join(be.destination, outname)
		if fd, err = os.Create(outfile); err != nil {
			return err
		}
		mw = io.MultiWriter(fd, h)
		pk = pack.NewTargzPacker(mw)
	}
	// Please keep order
	defer func() {
		if err == nil && h != nil {
			fmt.Fprintf(os.Stderr, "\x1b[34m%s  %s\x1b[0m\n", hex.EncodeToString(h.Sum(nil)), outname)
			fmt.Fprintf(os.Stderr, "bali create archive \x1b[32m%s\x1b[0m success\n", outname)
		}
	}()
	defer fd.Close()
	defer pk.Close()

	for _, b := range be.binaries {
		var rel string
		if rel, err = filepath.Rel(be.out, b); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "compress target \x1b[32m%s\x1b[0m\n", rel)
		if err = pk.AddFileEx(b, be.PathInArchive(rel), true); err != nil {
			return err
		}
	}
	for name, lnkName := range be.linkmap {
		nameInArchive := be.PathInArchive(name)
		fmt.Fprintf(os.Stderr, "compress link \x1b[32m%s --> %s\x1b[0m\n", nameInArchive, lnkName)
		if err = pk.AddTargetLink(nameInArchive, lnkName); err != nil {
			return err
		}
	}
	for _, f := range be.bm.Files {
		file := filepath.Join(be.workdir, f.Path)
		rel := filepath.Join(f.Destination, f.Base())
		fmt.Fprintf(os.Stderr, "compress profile \x1b[32m%s\x1b[0m\n", f.Path)
		if err = pk.AddFileEx(file, be.PathInArchive(rel), f.Executable); err != nil {
			return err
		}
	}
	return nil
}

// PackWin todo
func (be *Executor) PackWin() error {
	fmt.Fprintf(os.Stderr, "Windows installation package is not yet supported\n")
	return nil
}

// PackUNIX todo
func (be *Executor) PackUNIX() error {
	var outfilename string
	if be.withoutVersion {
		outfilename = base.StrCat(be.bm.Name, "-", be.target, "-", be.arch, ".sh")
	} else {
		outfilename = base.StrCat(be.bm.Name, "-", be.target, "-", be.arch, "-", be.bm.Version, ".sh")
	}
	outfile := filepath.Join(be.destination, outfilename)
	hashfd, err := pack.OpenHashableFile(outfile)
	if err != nil {
		return err
	}
	// Keep order
	defer func() {
		if err == nil {
			hashfd.Hashsum(outfilename)
			fmt.Fprintf(os.Stderr, "create \x1b[32m%s\x1b[0m done\n", outfile)
			fmt.Fprintf(os.Stderr, "Your can run '\x1b[32m./%s --prefix=/path/to/%s\x1b[0m' to install %s\n", outfilename, be.bm.Name, be.bm.Name)
			fmt.Fprintf(os.Stderr, "bali create package \x1b[32m%s\x1b[0m success\n", outfilename)
		}
	}()
	defer hashfd.Close()
	pk := pack.NewTargzPacker(hashfd)
	defer pk.Close()
	var rw pack.RespondWriter
	// bali post install script
	if len(be.bm.Respond) != 0 {
		if !base.PathExists(be.bm.Respond) {
			return base.ErrorCat("respond file ", be.bm.Respond, " not found")
		}
		if err = pk.AddFileEx(be.bm.Respond, "bali_post_install", true); err != nil {
			return err
		}
	} else if !be.norename {
		if err = rw.Initialize(); err != nil {
			return err
		}
		if err = rw.WriteBase(); err != nil {
			return err
		}
	}

	for _, s := range be.binaries {
		var rel string
		if rel, err = filepath.Rel(be.out, s); err != nil {
			_ = rw.Close()
			return err
		}
		fmt.Fprintf(os.Stderr, "compress target \x1b[32m%s\x1b[0m\n", rel)
		nameInArchive := be.PathInArchive(rel)
		if !be.norename {
			nameInArchive = base.StrCat(nameInArchive, ".new")
		}
		if err = pk.AddFileEx(s, nameInArchive, true); err != nil {
			_ = rw.Close()
			return err
		}
		DbgPrint("Add target %s", rel)
		_ = rw.AddTarget(nameInArchive)
	}
	for name, lnkName := range be.linkmap {
		nameInArchive := be.PathInArchive(name)
		fmt.Fprintf(os.Stderr, "compress link \x1b[32m%s --> %s\x1b[0m\n", nameInArchive, lnkName)
		if err = pk.AddTargetLink(nameInArchive, lnkName); err != nil {
			_ = rw.Close()
			return err
		}
	}
	for _, f := range be.bm.Files {
		file := filepath.Join(be.workdir, f.Path)
		rel := filepath.Join(f.Destination, f.Base())
		fmt.Fprintf(os.Stderr, "compress profile \x1b[32m%s\x1b[0m\n", rel)
		if be.norename || f.NoRename {
			if err = pk.AddFileEx(file, be.PathInArchive(rel), f.Executable); err != nil {
				_ = rw.Close()
				return err
			}
			DbgPrint("Add profile %s (no rename)", f.Path)
		} else {
			nameInArchive := base.StrCat(be.PathInArchive(rel), ".template")
			if err = pk.AddFileEx(file, nameInArchive, f.Executable); err != nil {
				_ = rw.Close()
				return err
			}
			DbgPrint("Add profile %s", f.Path)
			_ = rw.AddProfile(nameInArchive)
		}
	}
	_ = rw.Close()
	if len(rw.Path) != 0 {
		DbgPrint("Add post install script %s", rw.Path)
		if err = pk.AddFileEx(rw.Path, pack.RespondName, true); err != nil {
			return err
		}
	}
	return nil
}

// Pack todo
func (be *Executor) Pack() error {
	if !base.PathDirExists(be.destination) {
		if err := os.MkdirAll(be.destination, 0755); err != nil {
			return err
		}
	}
	return be.PackUNIX()
}
