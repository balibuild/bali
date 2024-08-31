package barrow

import (
	"context"
	"fmt"
	"time"

	"github.com/balibuild/bali/v3/module/ar"
)

var (
	archToDebain = map[string]string{
		"386":      "i386",
		"arm64":    "arm64",
		"arm5":     "armel",
		"arm6":     "armhf",
		"arm7":     "armhf",
		"mips64le": "mips64el",
		"mipsle":   "mipsel",
		"ppc64le":  "ppc64el",
		"s390":     "s390x",
	}
)

func debianArchGuard(arch string) string {
	if a, ok := archToDebain[arch]; ok {
		return a
	}
	return arch
}

func addArFile(w ar.Writer, name string, body []byte, date time.Time) error {
	header := &ar.Header{
		Name:    ToNixPath(name),
		Size:    int64(len(body)),
		Mode:    0o644,
		ModTime: date,
	}
	if err := w.WriteHeader(header); err != nil {
		return fmt.Errorf("cannot write file header: %w", err)
	}
	_, err := w.Write(body)
	return err
}

func (b *BarrowCtx) deb(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
