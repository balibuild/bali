// +build windows

package main

import (
	"io"

	"github.com/fcharlie/buna/debug/pe"
)

func overlayOffset(r io.ReaderAt) (int64, error) {
	file, err := pe.NewFile(r)
	if err != nil {
		return 0, err
	}
	return file.OverlayOffset, nil
}
