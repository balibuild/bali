// +build !windows

package inno

import (
	"os/exec"
)

func init() {
	innoExe, err := exec.LookPath("iscc")
	if err == nil {
		innoExePath = innoExe
		return
	}
}
