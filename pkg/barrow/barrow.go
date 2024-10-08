package barrow

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type BarrowCtx struct {
	CWD         string
	Out         string
	Target      string
	Arch        string
	Release     string
	Destination string
	Pack        []string // supported: zip, tar, sh, rpm
	Compression string
	Verbose     bool
	extraEnv    map[string]string
	environ     []string
	// TODO signature
}

func (b *BarrowCtx) DbgPrint(format string, a ...any) {
	if !b.Verbose {
		return
	}
	message := fmt.Sprintf(format, a...)
	message = strings.TrimRightFunc(message, unicode.IsSpace)
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		fmt.Fprintf(os.Stderr, "\x1b[38;2;255;215;0m* %s\x1b[0m\n", line)
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

func (b *BarrowCtx) binaryName(name string) string {
	if b.Target == "windows" {
		return name + ".exe"
	}
	return name
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
	b.extraEnv["BUILD_TARGET"] = b.Target
	b.extraEnv["BUILD_ARCH"] = b.Arch
	if err := b.resolveGit(ctx); err != nil {
		return err
	}
	if runID, ok := os.LookupEnv("GITHUB_RUN_ID"); ok {
		b.extraEnv["BUILD_RELEASE"] = runID
	}
	if len(b.Release) == 0 {
		b.Release = b.Getenv("BUILD_RELEASE")
	}
	t := time.Now()
	b.extraEnv["BUILD_TIME"] = t.Format(time.RFC3339)
	b.extraEnv["BUILD_YEAR"] = strconv.Itoa(t.Year())
	b.makeEnv()
	return nil
}

func (b *BarrowCtx) debugEnv() {
	lines := make([]string, 0, len(b.extraEnv))
	for k, v := range b.extraEnv {
		lines = append(lines, k+"="+v)
	}
	slices.Sort(lines)
	for _, line := range lines {
		fmt.Fprintf(os.Stderr, "\x1b[38;2;255;215;0m* env: %s\x1b[0m\n", line)
	}
}

func (b *BarrowCtx) Run(ctx context.Context) error {
	p, err := b.LoadPackage(b.CWD)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse package metadata error: %v\n", err)
		return err
	}
	b.DbgPrint("Building %s version: %s target: %s arch: %s", p.Name, p.Version, b.Target, b.Arch)
	b.extraEnv["BUILD_VERSION"] = p.Version

	if b.Verbose {
		b.debugEnv()
	}

	for _, item := range p.Include {
		if err := b.apply(item); err != nil {
			fmt.Fprintf(os.Stderr, "apply item %s error: %v\n", item.Path, err)
			return err
		}
	}
	// compile crates
	crates := make([]*Crate, 0, len(p.Crates))
	for _, location := range p.Crates {
		crate, err := b.compile(ctx, location)
		if err != nil {
			return err
		}
		crates = append(crates, crate)
	}
	if len(b.Pack) == 0 {
		return nil
	}
	for _, pack := range b.Pack {
		switch strings.ToLower(pack) {
		case "zip":
			if err := b.zip(ctx, p, crates); err != nil {
				fmt.Fprintf(os.Stderr, "bali create zip package error: %v\n", err)
				return err
			}
		case "rpm":
			if err := b.rpm(ctx, p, crates); err != nil {
				fmt.Fprintf(os.Stderr, "bali create rpm package error: %v\n", err)
				return err
			}
		case "sh":
			if err := b.sh(ctx, p, crates); err != nil {
				fmt.Fprintf(os.Stderr, "bali create sh package error: %v\n", err)
				return err
			}
		case "tar":
			if err := b.tar(ctx, p, crates); err != nil {
				fmt.Fprintf(os.Stderr, "bali create tar package error: %v\n", err)
				return err
			}
		case "deb":
			if err := b.deb(ctx, p, crates); err != nil {
				fmt.Fprintf(os.Stderr, "bali create deb package error: %v\n", err)
				return err
			}
		case "apk":
			if err := b.apk(ctx, p, crates); err != nil {
				fmt.Fprintf(os.Stderr, "bali create deb package error: %v\n", err)
				return err
			}
		case "arch":
			if err := b.archLinux(ctx, p, crates); err != nil {
				fmt.Fprintf(os.Stderr, "bali create deb package error: %v\n", err)
				return err
			}
		default:
			fmt.Fprintf(os.Stderr, "unsupported pack format '%s'\n", b.Pack)
			return fmt.Errorf("unsupported pack format '%s'", b.Pack)
		}
	}
	return nil
}

func (b *BarrowCtx) makeAlias(from, to string) error {
	if !filepath.IsAbs(to) {
		to = filepath.Join(b.Out, to)
	}
	if _, err := os.Lstat(to); err == nil {
		_ = os.Remove(to)
	}
	if err := os.Symlink(from, to); err != nil {
		fmt.Fprintf(os.Stderr, "create symlink error: %v\n", err)
		return err
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
		defer releaseFn() // remove it
	}
	b.DbgPrint("crate: %s\n", crate.Name)
	baseName := b.binaryName(crate.Name)
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
	stage("compile", "crate: %s version: %s for %s/%s", crate.Name, crate.Version, b.Target, b.Arch)
	status("%s", cmdStringsArgs(cmd))
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "compile %s error \x1b[31m%s\x1b[0m\n", crate.Name, err)
		return nil, err
	}
	crateDestination := filepath.Join(crate.Destination, baseName)
	crateFullPath := filepath.Join(b.Out, crateDestination)
	_ = os.MkdirAll(filepath.Dir(crateFullPath), 0755)
	if err := os.Rename(filepath.Join(crate.cwd, baseName), crateFullPath); err != nil {
		fmt.Fprintf(os.Stderr, "move out to dest error: %v\n", err)
		return nil, err
	}
	for _, a := range crate.Alias {
		aliasExpend := b.ExpandEnv(b.binaryName(a))
		stage("compile", "Link \x1b[38;02;39;199;173m%s\x1b[0m --> \x1b[38;02;39;199;173m%s\x1b[0m ", filepath.ToSlash(crateDestination), filepath.ToSlash(aliasExpend))
		if err := b.makeAlias(crateFullPath, aliasExpend); err != nil {
			return nil, err
		}
	}
	return crate, nil
}

func (b *BarrowCtx) Cleanup(force bool) error {
	p, err := b.LoadPackage(b.CWD)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse package metadata error: %v\n", err)
		return err
	}
	for _, item := range p.Include {
		if err := b.cleanupItem(item, force); err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[31mcleanup %s error: %v\x1b[0m\n", item.Path, err)
		}
	}
	for _, location := range p.Crates {
		if err := b.cleanupCrate(location); err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[31mcleanup %s error: %v\x1b[0m\n", location, err)
		}
	}
	if err := b.cleanupPackages(); err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[31mcleanup packages error: %v\x1b[0m\n", err)
	}
	return nil
}
