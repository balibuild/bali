package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/balibuild/bali/utilities"
)

// BaliExecutor

// BaliExecutor todo
type BaliExecutor struct {
	de          *utilities.Derivator
	target      string // os
	arch        string
	out         string
	workdir     string
	destination string
	makezip     bool
	makepack    bool
	norename    bool
	cleanup     bool
	environ     []string
	binaries    []string
	bm          Project
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

func resolveBranch(cwd string) string {
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = cwd
	if out, err := cmd.CombinedOutput(); err == nil {
		branch := strings.TrimSpace(strings.TrimPrefix(string(out), "refs/heads/"))
		DbgPrint("BUILD_BRANCH: '%s'", branch)
		return branch
	}
	return "None"
}

func resolveGoVersion() string {
	cmd := exec.Command("go", "version")
	if out, err := cmd.CombinedOutput(); err == nil {
		goversion := strings.TrimPrefix(strings.TrimSpace(string(out)), "go version ")
		DbgPrint("Go Version: '%s'", goversion)
		return goversion
	}
	return "None"
}

func resolveDistSupport(target, arch string) bool {
	cmd := exec.Command("go", "tool", "dist", "list")
	if out, err := cmd.CombinedOutput(); err == nil {
		str := utilities.StrCat(target, "/", arch)
		if strings.Contains(string(out), str) {
			return true
		}
	}
	return false
}

// TargetName todo
func (be *BaliExecutor) TargetName(name string) string {
	if be.norename {
		return name
	}
	return utilities.StrCat(name, ".new")
}

// FileName todo
func (be *BaliExecutor) FileName(file *File) string {
	if be.norename || file.NoRename {
		return file.Base()
	}
	return utilities.StrCat(file.Base(), ".template")
}

// Initialize todo
func (be *BaliExecutor) Initialize() error {
	bali := filepath.Join(be.workdir, "bali.json")
	if err := LoadMetadata(bali, &be.bm); err != nil {
		return err
	}
	if len(be.bm.Version) == 0 {
		be.bm.Version = "0.0.1"
	}
	be.de = utilities.NewDerivator()
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
		return utilities.ErrorCat("unsupported GOOS/GOARCH pair ", be.target, "/", be.arch)
	}
	_ = be.de.Append("BUILD_COMMIT", resolveBuildID(be.workdir))
	_ = be.de.Append("BUILD_BRANCH", resolveBranch(be.workdir))
	_ = be.de.Append("BUILD_GOVERSION", resolveGoVersion())
	_ = be.de.Append("BUILD_VERSION", "0.0.1")
	t := time.Now()
	_ = be.de.Append("BUILD_TIME", t.Format(time.RFC3339))
	osenv := os.Environ()
	be.environ = make([]string, 0, len(osenv)+3)
	for _, e := range osenv {
		// remove old env
		if !strings.HasPrefix(e, "GOOS=") && !strings.HasPrefix(e, "GOARCH=") {
			be.environ = append(be.environ, e)
		}
	}
	be.environ = append(be.environ, utilities.StrCat("GOOS=", be.target))
	be.environ = append(be.environ, utilities.StrCat("GOARCH=", be.arch))
	return nil
}
