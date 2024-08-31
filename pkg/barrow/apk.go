package barrow

var (
	// https://wiki.alpinelinux.org/wiki/Architecture
	archToAlpine = map[string]string{
		"386":     "x86",
		"amd64":   "x86_64",
		"arm64":   "aarch64",
		"arm6":    "armhf",
		"arm7":    "armv7",
		"ppc64le": "ppc64le",
		"s390":    "s390x",
	}
)

func alpineArchGuard(arch string) string {
	if a, ok := archToDebain[arch]; ok {
		return a
	}
	return arch
}
