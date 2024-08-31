package barrow

// ArchLinux use pacman. msys2 also

var (
	archToArchLinux = map[string]string{
		"all":   "any",
		"amd64": "x86_64",
		"386":   "i686",
		"arm64": "aarch64",
		"arm7":  "armv7h",
		"arm6":  "armv6h",
		"arm5":  "arm",
	}
)

func archLinuxArchGuard(arch string) string {
	if a, ok := archToDebain[arch]; ok {
		return a
	}
	return arch
}
