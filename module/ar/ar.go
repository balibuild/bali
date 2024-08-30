// Package ar provides both a reader and a writer for Unix ar archives. The
// writer allows for optional automatic file size determination. This way,
// entries can be written without knowing the final file size beforehand, for
// example when adding files that are being compressed simultaneously.
package ar

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"strings"
	"time"
)

const (
	// GlobalHeader contains the magic bytes that are written at the beginning
	// of an ar file (also known as armag).
	GlobalHeader = "!<arch>\n"
	// ThinArchiveGlobalHeader contains the magic bytes that are written at the
	// beginning of a thin ar archive, which is currently not supported.
	ThinArchiveGlobalHeader = "!<thin>\n"
	// HeaderTerminator contains the byte sequence with which file headers are
	// terminated (also known as ar_fmag).
	HeaderTerminator = "`\n"
	// HeaderSize holds the size of an ar header in bytes.
	HeaderSize = 60
	// UnknownSize signals that an ar entry's size is not known before writing
	// it. If Header.Size holds UnknownSize, Writer will automatically determine
	// the size and correct the header's size value when finalizing the entry.
	// If the underlying writer is not an *os.File (or does not implement Write,
	// WriteAt and Seek), this requires the Writer to buffer the file content in
	// memory.
	UnknownSize = math.MaxInt64
)

var (
	// ErrWriteTooLong is returned when more bytes are written than advertized
	// in the file header. ErrWriteTooLong is not returned in auto-correcting
	// mode (when Header.Size == UnknownSize) where the header size is
	// retroactively corrected instead.
	ErrWriteTooLong = fmt.Errorf("write too long")

	// ErrWriteTooShort is returned when a new header is appended or if the file
	// is closed before writing as many bytes for the previous file entry as
	// advertized in the corresponding header. ErrWriteTooShort is not returned
	// in auto-correcting mode (when Header.Size == UnknownSize) where the
	// header size is retroactively corrected instead.
	ErrWriteTooShort = fmt.Errorf("write too short")

	// ErrInvalidGlobalHeader is returned by NewReader if the provided data does
	// not start with the correct ar global header.
	ErrInvalidGlobalHeader = fmt.Errorf("invalid global header")
)

const (
	nameFieldSize    = 16
	modTimeFieldSize = 12
	uidFieldSize     = 6
	gidFieldSize     = 6
	modeFiledSize    = 8
	sizeFieldSize    = 10

	maxSize = 9999999999

	padding                    = '\n'
	bsdExtendedFormatPrefix    = "#1/"
	gnuExtendedFormatNameTable = "//"
)

// Type specifies the ar archive variant with the corresponding extensions.
type Type int

var (
	// TypeBasic is a basic ar archive without any extensions that supports 16
	// character file names.
	TypeBasic Type = 0 //nolint:revive
	// TypeBSD is the ar archive type written by BSD's ar tool with support
	// for large file names.
	TypeBSD Type = 1
	// TypeGNU is the ar archive type written by GNU's or SYSTEM V's ar tool
	// with support for large file names.
	TypeGNU Type = 2
)

// Header holds an ar entry's metadata.
type Header struct {
	// Name holds the file name. If the name has more than 16 characters, it is
	// considered an extended name which will be written in BSD style.
	Name string
	// ModTime holds the time stamp of the last file modification.
	ModTime time.Time
	// UID holds the file owner's UID.
	UID int64
	// GID holds the fole owner's GID.
	GID int64
	// Mode holds the file's mode.
	Mode uint32
	// Size holds the file size of up to 9999999999 bytes. If Size is
	// UnknownSize, the file size will be automatically determined by the actual
	// bytes written for the entry.
	Size int64
}

func (h *Header) hasExtendedName() bool {
	return len(h.Name) > nameFieldSize || strings.Contains(h.Name, " ")
}

func (h *Header) extendedNameSize() int {
	return len(h.Name) + (4 - len(h.Name)%4) //nolint:gomnd
}

func (h *Header) extendedNameBytes() []byte {
	return append([]byte(h.Name), make([]byte, (4-len(h.Name)%4))...) //nolint:gomnd
}

// NewHeaderFromFile creates an ar header based on the stat information of a
// file.
func NewHeaderFromFile(fileName string) (*Header, error) {
	stat, err := os.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	return NewHeaderFromFileInfo(stat), nil
}

// NewHeaderFromFileInfo creates an ar header based on the provided fs.FileInfo
// as returned for examply by os.Stat.
func NewHeaderFromFileInfo(stat fs.FileInfo) *Header {
	return &Header{
		Name:    stat.Name(),
		ModTime: stat.ModTime(),
		Mode:    uint32(stat.Mode()),
		Size:    stat.Size(),
	}
}
