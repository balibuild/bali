// +build darwin

package main

import (
	"io"

	"github.com/fcharlie/buna/debug/macho"
)

func overlayOffset(r io.ReaderAt) (int64, error) {
	file, err := macho.NewFile(r)
	if err != nil {
		return 0, err
	}
	return int64(file.OverlayOffset), nil
}
