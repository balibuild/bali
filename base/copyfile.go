package base

import (
	"fmt"
	"io"
	"os"
)

// CopyFile todo
func CopyFile(src, dest string) error {
	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", st.Name(), st.Mode().String())
	}
	dst, err := os.Stat(dest)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !dst.Mode().IsRegular() {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dst.Name(), dst.Mode().String())
		}
		if os.SameFile(st, dst) {
			return nil
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, st.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
