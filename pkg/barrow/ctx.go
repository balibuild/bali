package barrow

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

type BarrowCtx struct {
	CWD      string
	Out      string
	Target   string
	Arch     string
	Release  string
	Pack     string // pack: zip, tgz, stgz,rpm
	Verbose  bool
	extraEnv map[string]string
	environ  []string
}

func (b *BarrowCtx) DbgPrint(format string, a ...any) {
	if !b.Verbose {
		return
	}
	message := fmt.Sprintf(format, a...)
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		fmt.Fprintf(os.Stderr, "\x1b[33m* %s\x1b[0m\n", strings.TrimRightFunc(line, unicode.IsSpace))
	}
}

func (b *BarrowCtx) Getenv(key string) string {
	if v, ok := b.extraEnv[key]; ok {
		return v
	}
	return os.Getenv(key)
}

func (b *BarrowCtx) ExpandEnv(s string) string {
	return os.Expand(s, b.Getenv)
}

func (b *BarrowCtx) LookupEnv(key string) (string, bool) {
	if v, ok := b.extraEnv[key]; ok {
		return v, true
	}
	return os.LookupEnv(key)
}

func resolveGoVersion(ctx context.Context) (version string, host string) {
	cmd := exec.CommandContext(ctx, "go", "version")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	line := strings.TrimRightFunc(string(out), unicode.IsSpace)
	sv := strings.Split(line, " ")
	if len(sv) < 4 {
		return
	}
	version = strings.TrimPrefix(sv[2], "go")
	host = sv[3]
	return
}

func isDistSupported(ctx context.Context, target, arch string) bool {
	distName := target + "/" + arch
	cmd := exec.CommandContext(ctx, "go", "tool", "dist", "list")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "check dist is supported error: %v\n", err)
		return false
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "check dist is supported error: %v\n", err)
		return false
	}
	defer cmd.Wait()
	br := bufio.NewScanner(stdout)
	for br.Scan() {
		line := strings.TrimSpace(br.Text())
		if line == distName {
			return true
		}
	}
	return false
}

func (b *BarrowCtx) makeEnv() {
	originEnv := os.Environ()
	b.environ = make([]string, 0, len(originEnv))
	for _, e := range originEnv {
		k, _, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		if _, ok = b.extraEnv[k]; ok {
			continue
		}
		b.environ = append(b.environ, e)
	}
	for k, v := range b.extraEnv {
		b.environ = append(b.environ, k+"="+v)
	}
}

func (b *BarrowCtx) Initialize(ctx context.Context) error {
	version, host := resolveGoVersion(ctx)
	if !isDistSupported(ctx, b.Target, b.Arch) {
		fmt.Fprintf(os.Stderr, "golang %s (dist: %s) not support: %s/%s\n", version, host, b.Target, b.Arch)
		return errors.New("dist not supported")
	}
	b.extraEnv = make(map[string]string)
	b.extraEnv["BUILD_GOVERSION"] = version
	b.extraEnv["BUILD_HOST"] = host
	b.extraEnv["GOOS"] = b.Target
	b.extraEnv["GOARCH"] = b.Arch
	if err := b.resolveGit(ctx); err != nil {
		return err
	}
	if runID, ok := os.LookupEnv("GITHUB_RUN_ID"); ok {
		b.extraEnv["BUILD_RELEASE"] = runID
	}
	if len(b.Release) == 0 {
		b.Release = b.Getenv("BUILD_RELEASE")
	}
	b.extraEnv["BUIlD_TIME"] = time.Now().Format(time.RFC3339)
	b.makeEnv()
	return nil
}

func (b *BarrowCtx) Run(ctx context.Context) error {
	p, err := LoadPackage(b.CWD)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse package metadata error: %v\n", err)
		return err
	}
	b.DbgPrint("load %s version: %s done", p.Name, p.Version)
	b.extraEnv["BUILD_VERSION"] = p.Version

	crates := make([]*Crate, 0, len(p.Crates))
	for _, c := range p.Crates {
		crate, err := b.compile(ctx, c)
		if err != nil {
			return err
		}
		crates = append(crates, crate)
	}
	fmt.Fprintf(os.Stderr, "crates: %d\n", len(crates))
	switch strings.ToLower(b.Pack) {
	case "zip":
	case "rpm":
	case "sh", "stgz":
	case "tar.gz", "tgz":
	case "":
		return nil
	default:
		return errors.New("bad format")
	}
	return nil
}

func (b *BarrowCtx) compile(ctx context.Context, location string) (*Crate, error) {
	crate, err := b.LoadCrate(location)
	if err != nil {
		return nil, err
	}
	releaseFn, err := b.MakeResources(crate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "crate: %s build resources error \x1b[31m%s\x1b[0m\n", crate.Name, err)
		return nil, err
	}
	if releaseFn != nil {
		releaseFn()
	}
	b.DbgPrint("crate: %s\n", crate.Name)
	baseName := crate.baseName(b.Target)
	psArgs := make([]string, 0, 8)
	psArgs = append(psArgs, "build", "-o", baseName)
	for _, flag := range crate.GoFlags {
		psArgs = append(psArgs, b.ExpandEnv(flag))
	}
	cmd := exec.CommandContext(ctx, "go", psArgs...)
	cmd.Dir = crate.cwd
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = b.environ
	fmt.Fprintf(os.Stderr, "go compile \x1b[32m%s\x1b[0m version: \x1b[32m%s\x1b[0m\n", crate.Name, crate.Version)
	fmt.Fprintf(os.Stderr, "\x1b[34m%s\x1b[0m\n", cmd.String())
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "compile %s error \x1b[31m%s\x1b[0m\n", crate.Name, err)
		return nil, err
	}
	destTo := filepath.Join(b.Out, crate.Destination, baseName)
	_ = os.MkdirAll(filepath.Dir(destTo), 0755)
	if err := os.Rename(filepath.Join(crate.cwd, baseName), destTo); err != nil {
		fmt.Fprintf(os.Stderr, "move out to dest error: %v\n", err)
		return nil, err
	}
	return crate, nil
}
