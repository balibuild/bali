package cpio_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/balibuild/bali/v2/cpio"
)

func store(w *cpio.Writer, fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	hdr, err := cpio.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	if err := w.WriteHeader(hdr); err != nil {
		return err
	}
	if !fi.IsDir() {
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
	}
	return err
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	w := cpio.NewWriter(&buf)
	if err := store(w, "testdata/etc"); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := store(w, "testdata/etc/hosts"); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
