package ar

import (
	"bytes"
	"fmt"
	"io"
)

type defaultWriter struct {
	w                   io.Writer
	currentHeader       *Header
	currentBuffer       *bytes.Buffer
	emittedGlobalHeader bool
	remainingBytes      int64
}

var _ Writer = &defaultWriter{}

func (w *defaultWriter) WriteHeader(hdr *Header) error {
	// emit a global header before writing the first header
	if !w.emittedGlobalHeader {
		err := writeGlobalHeader(w.w)
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

	if hdr.Size == UnknownSize { // auto-correcting (buffered) mode
		// prepare a buffer for the data and flush it later in the next
		// finalizeEntry call together with the corrected header
		w.currentBuffer = &bytes.Buffer{}

		return nil
	}

	// in direct mode the file size has to be tracked but no buffer is needed
	w.currentBuffer = nil
	w.remainingBytes = hdr.Size

	return writeHeader(w.w, w.currentHeader)
}

func (w *defaultWriter) Write(data []byte) (int, error) {
	if w.currentHeader == nil {
		return 0, fmt.Errorf("writing data without header")
	}

	// in auto-correcting mode the data is buffered until the size is known
	if w.currentHeader.Size == UnknownSize {
		return w.currentBuffer.Write(data)
	}

	if int64(len(data)) > w.remainingBytes {
		return 0, fmt.Errorf(
			"writing %d bytes when only %d bytes remain for file %s of size %d: %w",
			len(data), w.remainingBytes, w.currentHeader.Name,
			w.currentHeader.Size, ErrWriteTooLong)
	}

	// in direct mode the header is already flushed and the data can be written
	// directly without buffering while keeping tracks of the number of bytes
	// written compared to the advertized file size
	n, err := w.w.Write(data)
	if err != nil {
		return 0, fmt.Errorf("writing to buffer: %w", err)
	}

	w.remainingBytes -= int64(n)

	return n, nil
}

func (w *defaultWriter) finalizeEntry() error {
	switch {
	case w.currentHeader == nil: // no entry yet
		return nil
	case w.currentHeader.Size == UnknownSize: // auto-correcting (buffered) mode
		// correct the file size
		w.currentHeader.Size = int64(w.currentBuffer.Len())

		// now we can flush the corrected header and buffer
		err := writeHeader(w.w, w.currentHeader)
		if err != nil {
			return fmt.Errorf("flushing previous header: %w", err)
		}

		_, err = io.Copy(w.w, w.currentBuffer)
		if err != nil {
			return fmt.Errorf("flushing previous file content: %w", err)
		}
	default: // direct mode
		// check if the advertized file size was written
		if w.remainingBytes > 0 {
			return fmt.Errorf(
				"%d bytes missing for file %s of size %d: %w",
				w.remainingBytes, w.currentHeader.Name,
				w.currentHeader.Size, ErrWriteTooShort)
		}
	}

	// add padding if necessary
	if w.currentHeader.Size%2 == 1 {
		_, err := w.w.Write([]byte{padding})
		if err != nil {
			return fmt.Errorf("add alignment byte: %w", err)
		}
	}

	// entry is finished, reset the state
	w.currentHeader = nil
	w.currentBuffer = nil

	return nil
}

func (w *defaultWriter) Close() error {
	return w.finalizeEntry()
}
