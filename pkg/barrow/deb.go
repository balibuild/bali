package barrow

import "context"

var (
	debianArchList = map[string]string{
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

func debianArchName(arch string) string {
	if a, ok := debianArchList[arch]; ok {
		return a
	}
	return arch
}

func (b *BarrowCtx) deb(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
