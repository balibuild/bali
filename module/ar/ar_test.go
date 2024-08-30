//nolint:errcheck
package ar

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"
)

func ExampleNewWriter() {
	// setup file to be copied into the ar archive
	f, _ := os.CreateTemp("", "ar_example")
	defer f.Close()
	defer os.RemoveAll(f.Name())

	var buffer bytes.Buffer

	arWriter := NewWriter(&buffer)

	stat, _ := os.Stat(f.Name())

	// write file header based on the file's stat
	err := arWriter.WriteHeader(NewHeaderFromFileInfo(stat))
	if err != nil {
		panic(err)
	}

	// write the file body (io.Copy uses arWriter.Write)
	_, err = io.Copy(arWriter, f)
	if err != nil {
		panic(err)
	}

	// make sure everything is written to the archive
	err = arWriter.Close()
	if err != nil {
		panic(err)
	}

	// print ar archive to stdout
	fmt.Println(buffer.String())
}

func ExampleWriter_WriteHeader_unknownSize() { //nolint:nosnakecase
	// setup temporary ar file
	f, _ := os.CreateTemp("", "ar_example")
	defer f.Close()
	defer os.RemoveAll(f.Name())

	data := []byte("compressible data: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	// write ar into a file such that the file size can be automatically
	// determined without buffering the file's data
	arWriter := NewWriter(f)

	// write a file header for the compressed data while signaling that the file
	// size needs to be automatically determined using the UnknownSize constant
	err := arWriter.WriteHeader(&Header{
		Name:    "data.gz",
		ModTime: time.Now(),
		Mode:    0o644,
		Size:    UnknownSize,
	})
	if err != nil {
		panic(err)
	}

	// write compressed data
	gzipWriter := gzip.NewWriter(arWriter)

	_, err = gzipWriter.Write(data)
	if err != nil {
		panic(err)
	}

	err = gzipWriter.Close()
	if err != nil {
		panic(err)
	}

	err = arWriter.Close()
	if err != nil {
		panic(err)
	}
}

func ExampleNewReader() {
	f, _ := os.Open("example.a")
	defer f.Close()

	arReader, err := NewReader(f)
	if err != nil {
		panic(err)
	}

	// read file header and setup reader to provide the file content
	header, err := arReader.Next()
	if err != nil {
		panic(err)
	}

	fmt.Println(header.Name + ":")

	// read the file content that belongs to the header that was returned by the
	// previous Next() call (io.ReadAll calls arReader.Read)
	content, err := io.ReadAll(arReader)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}
