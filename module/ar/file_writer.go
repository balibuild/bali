package ar

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

type fileWriter struct {
	f fileLike

	currentHeader          *Header
	currentEntrySize       int64
	currentEntrySizeOffset int64
	emittedGlobalHeader    bool
}

var _ Writer = &fileWriter{}

type fileLike interface {
	io.Seeker
	io.Writer
	io.WriterAt
}

var _ fileLike = &os.File{} // ensure *os.File implements fileLike

func (w *fileWriter) WriteHeader(hdr *Header) error {
	// emit a global header before writing the first header
	if !w.emittedGlobalHeader {
		err := writeGlobalHeader(w.f)
		if err != nil {
			return err
		}

		w.emittedGlobalHeader = true
	}

	// ensure that the previous entry is completely finished and padding as
	// applied
	err := w.finalizeEntry()
	if err != nil {
		return fmt.Errorf("finalize previous entry: %w", err)
	}

	w.currentHeader = hdr

	if hdr.Size == UnknownSize { // auto-correcting mode
		// calculate the offset of the size field in the header which has to be
		// corrected later
		headerOffset, err := w.f.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("record header offset: %w", err)
		}

		w.currentEntrySizeOffset = (headerOffset + nameFieldSize +
			modTimeFieldSize + uidFieldSize + gidFieldSize + modeFiledSize)
	}

	return writeHeader(w.f, hdr)
}

func (w *fileWriter) Write(data []byte) (int, error) {
	if w.currentHeader == nil {
		return 0, fmt.Errorf("writing data without header")
	}

	if w.currentHeader.Size != UnknownSize &&
		w.currentEntrySize+int64(len(data)) > w.currentHeader.Size {
		return 0, fmt.Errorf(
			"writing %d bytes when only %d bytes remain for file %s of size %d: %w",
			len(data), w.currentHeader.Size-w.currentEntrySize, w.currentHeader.Name,
			w.currentHeader.Size, ErrWriteTooLong)
	}

	n, err := w.f.Write(data)
	if err != nil {
		return n, err
	}

	w.currentEntrySize += int64(n)

	return n, nil
}

func (w *fileWriter) finalizeEntry() error {
	switch {
	case w.currentHeader == nil: // no entry yet
		return nil
	case w.currentHeader.Size == UnknownSize: // auto-correcting mode
		if w.currentEntrySizeOffset == 0 {
			return fmt.Errorf("cannot find header of the current entry")
		}

		if w.currentHeader.hasExtendedName() {
			// account for extended name that was inserted before the content
			w.currentEntrySize += int64(w.currentHeader.extendedNameSize())
		}

		newSize, err := expandToByteField(strconv.FormatInt(w.currentEntrySize, 10),
			sizeFieldSize)
		if err != nil {
			return fmt.Errorf("pack actual file size: %w", err)
		}

		_, err = w.f.WriteAt(newSize, w.currentEntrySizeOffset)
		if err != nil {
			return fmt.Errorf("correcting file size: %w", err)
		}
	default: // direct mode
		// check if the advertized file size was written
		if w.currentEntrySize != w.currentHeader.Size {
			return fmt.Errorf(
				"%d bytes missing for file %s of size %d: %w",
				w.currentHeader.Size-w.currentEntrySize, w.currentHeader.Name,
				w.currentHeader.Size, ErrWriteTooShort)
		}
	}

	if w.currentEntrySize%2 == 1 {
		_, err := w.f.Write([]byte{padding})
		if err != nil {
			return fmt.Errorf("add alignment byte: %w", err)
		}
	}

	w.currentHeader = nil
	w.currentEntrySize = 0
	w.currentEntrySizeOffset = 0

	return nil
}

func (w *fileWriter) Close() error {
	err := w.finalizeEntry()
	if err != nil {
		return err
	}

	f, ok := w.f.(*os.File)
	if !ok {
		return nil
	}

	return f.Sync()
}
