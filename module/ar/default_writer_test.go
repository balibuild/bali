package ar

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

func TestDefaultWriterAutoCorrect(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	aw := newDefaultWriter(t, buf)

	hdrOne := Header{
		Name: "foo",
		Size: UnknownSize,
	}
	fileOne := bytes.Repeat([]byte("X"), 3)

	err := aw.WriteHeader(&hdrOne)
	if err != nil {
		t.Fatalf("write first header: %v", err)
	}

	n, err := aw.Write(fileOne)
	if err != nil {
		t.Fatalf("write file one: %v", err)
	}

	if n != len(fileOne) {
		t.Fatalf("%d bytes were reported written for file one instead of %d",
			n, len(fileOne))
	}

	hdrTwo := Header{
		Name: "bar",
		Size: UnknownSize,
	}

	// file will be writing in two write calls
	fileTwo := append(bytes.Repeat([]byte("Y"), 3), bytes.Repeat([]byte("Z"), 3)...)

	err = aw.WriteHeader(&hdrTwo)
	if err != nil {
		t.Fatalf("write second header: %v", err)
	}

	n, err = aw.Write(fileTwo[:len(fileTwo)/2])
	if err != nil {
		t.Fatalf("first write for file two: %v", err)
	}

	if n != len(fileTwo)/2 {
		t.Fatalf("%d bytes were reported written for file two instead of %d",
			n, len(fileOne)/2)
	}

	n, err = aw.Write(fileTwo[len(fileTwo)/2:])
	if err != nil {
		t.Fatalf("second write for file two: %v", err)
	}

	if n != len(fileTwo)/2 {
		t.Fatalf("%d bytes were reported written for file two instead of %d",
			n, len(fileOne)/2)
	}

	err = aw.Close()
	if err != nil {
		t.Fatalf("closing ar file: %v", err)
	}

	reader, err := NewReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("create reader: %v", err)
	}

	readHdrOne, err := reader.Next()
	if err != nil {
		t.Fatalf("reading first file: %v", err)
	}

	if readHdrOne.Size != int64(len(fileOne)) {
		t.Fatalf("file one size was corrected to %d instead of %d",
			readHdrOne.Size, len(fileOne))
	}

	readFileOne := make([]byte, len(fileOne))

	n, err = reader.Read(readFileOne)
	if err != nil {
		t.Fatalf("reading file one: %v", err)
	}

	if n != len(fileOne) {
		t.Fatalf("%d bytes were read for file one instead of %d", n, len(fileOne))
	}

	if !bytes.Equal(readFileOne, fileOne) {
		t.Fatalf("file one does not match")
	}

	readHdrTwo, err := reader.Next()
	if err != nil {
		t.Fatalf("reading second file: %v", err)
	}

	if readHdrTwo.Size != int64(len(fileTwo)) {
		t.Fatalf("file two size was corrected to %d instead of %d",
			readHdrTwo.Size, len(fileTwo))
	}

	readFileTwo := make([]byte, len(fileTwo))

	n, err = reader.Read(readFileTwo)
	if err != nil {
		t.Fatalf("reading file two: %v", err)
	}

	if n != len(fileTwo) {
		t.Fatalf("%d bytes were read for file two instead of %d", n, len(fileTwo))
	}

	if !bytes.Equal(readFileTwo, fileTwo) {
		t.Fatalf("file two does not match")
	}
}

func TestDefaultReCreateEvenFile(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	w := newDefaultWriter(t, buf)

	err := w.WriteHeader(&Header{
		Name:    "first_even",
		ModTime: time.Unix(1664113056, 0),
		UID:     501,
		GID:     20,
		Mode:    0o644,
		Size:    16,
	})
	if err != nil {
		t.Fatalf("write first header: %v", err)
	}

	_, err = w.Write(readFile(t, "testdata/first_even"))
	if err != nil {
		t.Fatalf("write first file: %v", err)
	}

	err = w.WriteHeader(&Header{
		Name:    "second_even",
		ModTime: time.Unix(1664113074, 0),
		UID:     501,
		GID:     20,
		Mode:    0o644,
		Size:    10,
	})
	if err != nil {
		t.Fatalf("write first header: %v", err)
	}

	_, err = w.Write(readFile(t, "testdata/second_even"))
	if err != nil {
		t.Fatalf("write first file: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("close: %v", err)
	}

	expectedOutput := readFile(t, "testdata/even_file_sizes.a")

	if !bytes.Equal(buf.Bytes(), expectedOutput) {
		t.Fatalf("ar file content mismatch")
	}
}

func TestDefaulReCreateEvenFileUnknownSize(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	w := newDefaultWriter(t, buf)

	err := w.WriteHeader(&Header{
		Name:    "first_even",
		ModTime: time.Unix(1664113056, 0),
		UID:     501,
		GID:     20,
		Mode:    0o644,
		Size:    UnknownSize,
	})
	if err != nil {
		t.Fatalf("write first header: %v", err)
	}

	_, err = w.Write(readFile(t, "testdata/first_even"))
	if err != nil {
		t.Fatalf("write first file: %v", err)
	}

	err = w.WriteHeader(&Header{
		Name:    "second_even",
		ModTime: time.Unix(1664113074, 0),
		UID:     501,
		GID:     20,
		Mode:    0o644,
		Size:    UnknownSize,
	})
	if err != nil {
		t.Fatalf("write first header: %v", err)
	}

	_, err = w.Write(readFile(t, "testdata/second_even"))
	if err != nil {
		t.Fatalf("write first file: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("close: %v", err)
	}

	expectedOutput := readFile(t, "testdata/even_file_sizes.a")

	if !bytes.Equal(buf.Bytes(), expectedOutput) {
		t.Fatalf("ar file content mismatch")
	}
}

func TestDefaultWriteTooLong(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	w := newDefaultWriter(t, buf)

	err := w.WriteHeader(&Header{Size: 2})
	if err != nil {
		t.Fatalf("write header: %v", err)
	}

	_, err = w.Write([]byte("123"))
	if !errors.Is(err, ErrWriteTooLong) {
		t.Fatalf("write did not result in %q", ErrWriteTooLong.Error())
	}
}

func TestDefaultWriteTooShort(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	w := newDefaultWriter(t, buf)

	err := w.WriteHeader(&Header{Size: 2})
	if err != nil {
		t.Fatalf("write header: %v", err)
	}

	_, err = w.Write([]byte("1"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	err = w.WriteHeader(&Header{})
	if !errors.Is(err, ErrWriteTooShort) {
		t.Fatalf("write did not result in %q", ErrWriteTooShort.Error())
	}
}

func TestDefaultExtendedFileName(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	content := []byte("content\n")

	w := newDefaultWriter(t, buf)

	err := w.WriteHeader(&Header{
		Name:    "very_long_file_name_that_does_not_fit_into_name_field.txt",
		ModTime: time.Unix(1664483321, 0),
		UID:     501,
		GID:     20,
		Mode:    0o644,
		Size:    int64(len(content)),
	})
	if err != nil {
		t.Fatalf("write header: %v", err)
	}

	_, err = w.Write(content)
	if err != nil {
		t.Fatalf("write content: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("close: %v", err)
	}

	expected, err := os.ReadFile("testdata/long_filename_bsd.a")
	if err != nil {
		t.Fatalf("read expected ar file content: %v", err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Fatalf("ar file mismatch:\ngot:\n%q\nexpected:\n%q",
			buf.String(), string(expected))
	}
}

func TestDefaultExtendedFileNameUnknownSize(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	w := newDefaultWriter(t, buf)

	err := w.WriteHeader(&Header{
		Name:    "very_long_file_name_that_does_not_fit_into_name_field.txt",
		ModTime: time.Unix(1664483321, 0),
		UID:     501,
		GID:     20,
		Mode:    0o644,
		Size:    UnknownSize,
	})
	if err != nil {
		t.Fatalf("write header: %v", err)
	}

	_, err = w.Write([]byte("content\n"))
	if err != nil {
		t.Fatalf("write content: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("close: %v", err)
	}

	expected, err := os.ReadFile("testdata/long_filename_bsd.a")
	if err != nil {
		t.Fatalf("read expected ar file content: %v", err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Fatalf("ar file mismatch:\ngot:\n%q\nexpected:\n%q",
			buf.String(), string(expected))
	}
}

func newDefaultWriter(tb testing.TB, w io.Writer) *defaultWriter {
	tb.Helper()

	aw := NewWriter(w)

	fw, ok := aw.(*defaultWriter)
	if !ok {
		tb.Fatalf("NewWriter returned %T instead of a defaultWriter", aw)
	}

	return fw
}
