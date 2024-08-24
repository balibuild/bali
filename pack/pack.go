package pack

import "os"

// Builder todo
type Builder interface {
	Close() error
	AddFileEx(src, nameInArchive string, exerights bool) error
	AddFile(src, nameInArchive string) error
	AddTargetLink(nameInArchive, linkName string) error
}

func isSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}
