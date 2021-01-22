// +build !windows,!darwin

package main

import (
	"io"

	"github.com/fcharlie/buna/debug/elf"
)

func overlayOffset(r io.ReaderAt) (int64, error) {
	file, err := elf.NewFile(r)
	if err != nil {
		return 0, err
	}
	return int64(file.OverlayOffset), nil
}
