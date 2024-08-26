package barrow

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

func (b *BarrowCtx) resolveHEAD(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = b.CWD
	if out, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	if s, ok := os.LookupEnv("GITHUB_SHA"); ok {
		return s, nil
	}
	return "", os.ErrNotExist
}

func (b *BarrowCtx) resolveReferenceName(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD")
	cmd.Dir = b.CWD
	if out, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	if refname, ok := os.LookupEnv("GITHUB_REF"); ok {
		// Github support this
		return refname, nil
	}
	if tagName, ok := os.LookupEnv("GIT_BUILD_REF"); ok {
		// CODING support this
		return tagName, nil
	}
	return "", os.ErrNotExist
}

// git describe --tags --dirty
func (b *BarrowCtx) resolveDirtyTagName(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--dirty")
	cmd.Dir = b.CWD
	out, err := cmd.Output()
	if err != nil {
		// continue
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (b *BarrowCtx) resolveGit(ctx context.Context) error {
	if HEAD, err := b.resolveHEAD(ctx); err == nil {
		b.extraEnv["BUILD_COMMIT"] = HEAD
	}
	if n, err := b.resolveDirtyTagName(ctx); err == nil {
		b.extraEnv["BUILD_DIRTY_TAGNAME"] = n
	}
	if n, err := b.resolveReferenceName(ctx); err == nil {
		if branchName, ok := strings.CutPrefix(n, "refs/heads/"); ok {
			b.extraEnv["BUILD_BRANCH"] = branchName
		}
		b.extraEnv["BUILD_REFNAME"] = n
	}
	return nil
}
