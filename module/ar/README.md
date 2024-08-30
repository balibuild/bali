<p align="center">
  <h1 align="center"><b>ar</b></h1>
  <p align="center"><i>The Unix ar archive library for Go</i></p>
  <p align="center">
    <a href="https://pkg.go.dev/github.com/erikgeiser/ar"><img alt="Go Doc" src="https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge"></a>
    <a href="https://github.com/erikgeiser/ar/actions?workflow=Tests"><img alt="GitHub Action: Tests" src=" https://img.shields.io/github/actions/workflow/status/erikgeiser/ar/tests.yml?branch=main&label=Tests&style=for-the-badge"></a>
    </br>
    <a href="https://github.com/erikgeiser/ar/actions?workflow=Check"><img alt="GitHub Action: Check" src="https://img.shields.io/github/actions/workflow/status/erikgeiser/ar/check.yml?branch=main&label=Check&style=for-the-badge"></a>
    <a href="/LICENSE.md"><img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge"></a>
    <a href="https://goreportcard.com/report/github.com/erikgeiser/ar"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/erikgeiser/ar?style=for-the-badge"></a>

  </p>
</p>

This library provides a reader and a writer for Unix `ar` archives. The API
heavily inspired by the `tar` module from Go's standard library. The following
features set the library apart from other Go `ar` libraries such as
[github.com/blakesmith/ar](https://github.com/blakesmith/ar):

- **Automatic file size determination:** Add files without knowing their size
  beforehand. This is useful when compressing the files on the fly while writing
  them to `ar` archive by stacking multiple `io.Writer`.
- **Support for long file names:** The traditional `ar` format has a file name
  size limit of 16 characters. Multiple extensions have been created to work
  around this. `io.Writer` writes long archive names in BSD style and
  `io.Reader` supports BSD and System V/Gnu style.
- **Robust `io.Reader`/`io.Writer` APIs:** Files can be writting in multiple
  `Write` calls and they can be read in multiple `Read` calls.

## Reader

The `ar` Reader can read traditional, BSD-style and System V/Gnu-style archives.
Symbol lookup tables are currently not supported.

```go
arReader, err := ar.NewReader(file)
if err != nil {
  return err
}

// read file header and setup reader to provide the file content
header, err := arReader.Next()
if err != nil {
  return err
}
// read the file content that belongs to the header that was returned by the
// previous Next() call (io.ReadAll calls arReader.Read)
content, err := io.ReadAll(arReader)
if err != nil {
  return err
}
```

## Writer

The `ar` writer supports automatic file size determination. Consider the
following example, where a file is gzip compressed on the fly. Since the final
size of the compressed file is not known until after it is written, the `Size`
field of the `ar` header is set to `ar.UnknownSize` to enable automatic file
size determination.

```go
arWriter := NewWriter(file)

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

_, err = io.Copy(gzipWriter, largeFile)
if err != nil {
  return err
}

err = gzipWriter.Close()
if err != nil {
  return err
}

err = arWriter.Close()
if err != nil {
  return err
}
```

### Tips

When adding actual files to the `ar` archive, the header can be easily populated
using the helper functions `NewHeaderFromFile` and `NewHeaderFromFileInfo`.

### Caveats for automatic file size determination:

- If the underlying `io.Writer` is an `*os.File` or if it at least implements
  `io.WriterAt` and `io.Seek`, automatic file size determination has a negligible
  performance impact.
- If the underlying `io.Writer` is not an `*os.File` and if it does not
  implement `WriteAt` and `io.Seek`, the header and file content will be
  buffered in memory until the next `ar` entry is started or `Close` is called
  if automatic file size determination is enabled.

### Long file names

If a file name has more than 16 characters, it will be written in BSD-style
which can also be read by the Gnu/binutils `ar` tool.
