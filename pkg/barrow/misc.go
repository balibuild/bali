package barrow

import "os"

func nonEmpty(a string, dv string) string {
	if len(a) != 0 {
		return a
	}
	return dv
}

func isSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}
