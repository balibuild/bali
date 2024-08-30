package ar

import (
	"bytes"
	"os"
	"testing"
)

func TestNewWriterFile(t *testing.T) {
	t.Parallel()

	f := tempFile(t)

	w := NewWriter(f)

	fw, ok := w.(*fileWriter)
	if !ok {
		t.Fatalf("NewWriter returned %T instead of %T", w, fw)
	}
}

func TestNewWriterDefault(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	w := NewWriter(buf)

	dw, ok := w.(*defaultWriter)
	if !ok {
		t.Fatalf("NewWriter returned %T instead of %T", w, dw)
	}
}

func tempFile(tb testing.TB) *os.File {
	tb.Helper()

	f, err := os.CreateTemp("", "ar_test")
	if err != nil {
		tb.Fatalf("create temp file: %v", err)
	}

	tb.Cleanup(func() {
		err := f.Close()
		if err != nil {
			tb.Fatalf("close temp file: %v", err)
		}

		err = os.RemoveAll(f.Name())
		if err != nil {
			tb.Fatalf("remove temp file: %v", err)
		}
	})

	return f
}
