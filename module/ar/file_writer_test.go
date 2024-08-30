package ar

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

func TestFileWriterAutoCorrect(t *testing.T) {
	t.Parallel()

	f := tempFile(t)

	w := newFileWriter(t, f)

	hdrOne := Header{
		Name: "foo",
		Size: UnknownSize,
	}
	fileOne := bytes.Repeat([]byte("X"), 3)

	err := w.WriteHeader(&hdrOne)
	if err != nil {
		t.Fatalf("write first header: %v", err)
	}

	n, err := w.Write(fileOne)
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

	err = w.WriteHeader(&hdrTwo)
	if err != nil {
		t.Fatalf("write second header: %v", err)
	}

	n, err = w.Write(fileTwo[:len(fileTwo)/2])
	if err != nil {
		t.Fatalf("first write for file two: %v", err)
	}

	if n != len(fileTwo)/2 {
		t.Fatalf("%d bytes were reported written for file two instead of %d",
			n, len(fileOne)/2)
	}

	n, err = w.Write(fileTwo[len(fileTwo)/2:])
	if err != nil {
		t.Fatalf("second write for file two: %v", err)
	}

	if n != len(fileTwo)/2 {
		t.Fatalf("%d bytes were reported written for file two instead of %d",
			n, len(fileOne)/2)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("closing ar file: %v", err)
	}

	err = f.Sync()
	if err != nil {
		t.Fatalf("sync file: %v", err)
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("seek file start: %v", err)
	}

	reader, err := NewReader(f)
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

func TestFileReCreateEvenFile(t *testing.T) {
	t.Parallel()

	f := tempFile(t)

	w := newFileWriter(t, f)

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

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("seek: %v", err)
	}

	output, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read ar file: %v", err)
	}

	expectedOutput := readFile(t, "testdata/even_file_sizes.a")

	if !bytes.Equal(output, expectedOutput) {
		t.Fatalf("ar file content mismatch")
	}
}

func TestFileReCreateEvenFileUnknownSize(t *testing.T) {
	t.Parallel()

	f := tempFile(t)

	w := newFileWriter(t, f)

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

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("seek: %v", err)
	}

	output, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read ar file: %v", err)
	}

	expectedOutput := readFile(t, "testdata/even_file_sizes.a")

	if !bytes.Equal(output, expectedOutput) {
		t.Fatalf("ar file content mismatch")
	}
}

func TestFileWriteTooLong(t *testing.T) {
	t.Parallel()

	w := newFileWriter(t, tempFile(t))

	err := w.WriteHeader(&Header{Size: 2})
	if err != nil {
		t.Fatalf("write header: %v", err)
	}

	_, err = w.Write([]byte("123"))
	if !errors.Is(err, ErrWriteTooLong) {
		t.Fatalf("write did not result in %q", ErrWriteTooLong.Error())
	}
}

func TestFileWriteTooShort(t *testing.T) {
	t.Parallel()

	w := newFileWriter(t, tempFile(t))

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

func TestFileExtendedFileName(t *testing.T) {
	t.Parallel()

	content := []byte("content\n")

	f := tempFile(t)
	w := newFileWriter(t, f)

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

	result := fileContent(t, f)

	expected, err := os.ReadFile("testdata/long_filename_bsd.a")
	if err != nil {
		t.Fatalf("read expected ar file content: %v", err)
	}

	if !bytes.Equal(expected, result) {
		t.Fatalf("ar file mismatch:\ngot:\n%q\nexpected:\n%q",
			string(result), string(expected))
	}
}

func TestFileExtendedFileNameUnknownSize(t *testing.T) {
	t.Parallel()

	f := tempFile(t)
	w := newFileWriter(t, f)

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

	result := fileContent(t, f)

	expected, err := os.ReadFile("testdata/long_filename_bsd.a")
	if err != nil {
		t.Fatalf("read expected ar file content: %v", err)
	}

	if !bytes.Equal(expected, result) {
		t.Fatalf("ar file mismatch:\ngot:\n%q\nexpected:\n%q",
			string(result), string(expected))
	}
}

func newFileWriter(tb testing.TB, w io.Writer) *fileWriter {
	tb.Helper()

	aw := NewWriter(w)

	fw, ok := aw.(*fileWriter)
	if !ok {
		tb.Fatalf("NewWriter returned %T instead of a fileWriter", aw)
	}

	return fw
}

func fileContent(tb testing.TB, f *os.File) []byte {
	tb.Helper()

	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		tb.Fatalf("seek: %v", err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		tb.Fatalf("read file content: %v", err)
	}

	return content
}
