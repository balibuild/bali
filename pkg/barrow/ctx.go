package barrow

import (
	"context"
	"fmt"
	"os"
	"strings"
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

func (b *BarrowCtx) Initialize(ctx context.Context) error {
	b.extraEnv = make(map[string]string)
	if err := b.resolveGit(ctx); err != nil {
		return err
	}
	if runID, ok := os.LookupEnv("GITHUB_RUN_ID"); ok {
		b.extraEnv["BUILD_RELEASE"] = runID
	}

	// fill all values
	if len(b.Release) == 0 {
		b.Release = b.Getenv("BUILD_RELEASE")
	}
	return nil
}

func (b *BarrowCtx) Run(ctx context.Context) error {
	return nil
}
