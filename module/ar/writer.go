package ar

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Writer allows for sequential writing of an ar file by consecutively writing
// each file's header using WriteHeader and body using one or more calls to
// Write. Remember to call Close after finishing the last entry.
type Writer interface {
	// WriteHeader writes the file entry header and if necessary the global ar
	// header. By setting Header.Size == UnknownSize, the file size will be
	// auto-corrected. File names longer than 16 characters will be written
	// using BSD file name extension.
	WriteHeader(*Header) error

	// Write writes the actual file content corresponding the file header that
	// was previously written using WriteHeader. An ar file can be written in
	// multiple consecutive Write calls.
	Write([]byte) (int, error)

	// Close ensures that the complete ar content is flushed.
	Close() error
}

// NewWriter creates a new ar Writer that supports optional automatic file size
// correction (see Header.Size). If the provided writer is not an *os.File (or
// does not implement Seek, Write and WriteAt), auto-correcting the file size
// requires Writer to buffer the file content in memory.
func NewWriter(w io.Writer) Writer {
	f, ok := w.(*os.File)
	if ok {
		return &fileWriter{f: f}
	}

	return &defaultWriter{w: w}
}

func writeGlobalHeader(w io.Writer) error {
	_, err := w.Write([]byte(GlobalHeader))

	return err
}

func writeHeader(w io.Writer, hdr *Header) error {
	isExtendedName := hdr.hasExtendedName()

	name := hdr.Name
	if isExtendedName {
		// write BSD style placeholder for extended name
		name = bsdExtendedFormatPrefix + strconv.Itoa(
			hdr.extendedNameSize())
	}

	err := packString(w, name, nameFieldSize)
	if err != nil {
		return fmt.Errorf("write file name: %w", err)
	}

	err = packUint64(w, hdr.ModTime.Unix(), modTimeFieldSize)
	if err != nil {
		return fmt.Errorf("write mod time: %w", err)
	}

	err = packUint64(w, hdr.UID, uidFieldSize)
	if err != nil {
		return fmt.Errorf("write UID: %w", err)
	}

	err = packUint64(w, hdr.GID, gidFieldSize)
	if err != nil {
		return fmt.Errorf("write GID: %w", err)
	}

	err = packOctal(w, hdr.Mode, modeFiledSize)
	if err != nil {
		return fmt.Errorf("write file mode: %w", err)
	}

	// write a valid size placeholder value in auto-correct mode
	size := hdr.Size
	if size == UnknownSize {
		// extended file size will be accounted for in finalizeEntry
		size = maxSize
	} else if isExtendedName {
		size += int64(hdr.extendedNameSize())
	}

	err = packUint64(w, size, sizeFieldSize)
	if err != nil {
		return fmt.Errorf("write file size placeholder: %w", err)
	}

	err = packString(w, HeaderTerminator, len(HeaderTerminator))
	if err != nil {
		return fmt.Errorf("finishing header: %w", err)
	}

	if isExtendedName {
		// add extended name at the beginning of the content
		_, err = w.Write(hdr.extendedNameBytes())
		if err != nil {
			return fmt.Errorf("write BSD extended file name: %w", err)
		}
	}

	return nil
}

func packUint64(w io.Writer, value int64, fieldWidth int) error {
	return packString(w, strconv.FormatInt(value, 10), fieldWidth)
}

func packOctal(w io.Writer, value uint32, fieldWidth int) error {
	return packString(w, "100"+strconv.FormatUint(uint64(value), 8), fieldWidth)
}

func packString(w io.Writer, s string, fieldWidth int) error {
	data, err := expandToByteField(s, fieldWidth)
	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

func expandToByteField(s string, fieldWidth int) ([]byte, error) {
	switch {
	case len(s) < fieldWidth:
		return []byte(s + strings.Repeat(" ", fieldWidth-len(s))), nil
	case len(s) == fieldWidth:
		return []byte(s), nil
	default:
		return nil, fmt.Errorf("%d byte value %q is too large for %d byte field",
			len(s), s, fieldWidth)
	}
}
