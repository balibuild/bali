package ar

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Reader allows to sequentially read ar files by first reading an entry's
// header using Next and then reading the full file content using Read.
type Reader struct {
	// DisableBSDExtensions disables parsing the BSD file format extensions. If
	// this is disabled, the file names and data of ar files written with BSD's
	// or macOS's ar program will be corrupt if a file name is longer than 16
	// characters.
	DisableBSDExtensions bool

	// DisableGnuExtensions disables parsing the System V/Gnu file format. If
	// this is disabled, the file names and data of ar files written with the
	// Gnu/binutils ar program will be corrupt if a file name is longer than 16
	// characters.
	DisableGnuExtensions bool

	r             io.Reader
	remainingSize int64
	expectPadding bool
	headerBuffer  []byte
	gnuNameBuffer []byte
}

// NewReader creates a new ar file Reader.
func NewReader(r io.Reader) (*Reader, error) {
	// consume and check the global header
	globalHeaderBuffer := make([]byte, len(GlobalHeader))

	_, err := io.ReadFull(r, globalHeaderBuffer)
	if err != nil {
		return nil, fmt.Errorf("read global header: %w", err)
	}

	if !bytes.Equal(globalHeaderBuffer, []byte(GlobalHeader)) {
		return nil, fmt.Errorf("ar file starts with %q instead of %q: %w",
			string(globalHeaderBuffer), GlobalHeader, ErrInvalidGlobalHeader)
	}

	return &Reader{r: r, headerBuffer: make([]byte, HeaderSize)}, nil
}

// Next returns the header of the next file entry in the ar file and enables
// reading the corresponding body in subsequent Read calls. If Next is called
// before the previous file is completely read, the remaining bytes of the
// previous file will be skipped automatically.
func (r *Reader) Next() (*Header, error) {
	// skip unread bytes of previous entry and padding if necessary
	if r.expectPadding || r.remainingSize > 0 {
		_, err := io.CopyN(io.Discard,
			r.r, r.remainingSize+boolToInt64(r.expectPadding))
		if err != nil {
			return nil, fmt.Errorf("skip to next header: %w", err)
		}
	}

	var hdr Header

	err := r.parseHeader(&hdr)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	r.remainingSize = hdr.Size

	// anounce padding
	if hdr.Size%2 != 0 {
		r.expectPadding = true
	}

	return &hdr, nil
}

// Read reads the file content of the entry whose header was previously read using Next.
func (r *Reader) Read(buffer []byte) (n int, err error) {
	if r.remainingSize == 0 {
		return 0, io.EOF
	}

	if int64(len(buffer)) > r.remainingSize {
		buffer = buffer[:r.remainingSize]
	}

	n, err = r.r.Read(buffer)
	if err != nil {
		return n, err
	}

	r.remainingSize -= int64(n)

	return n, err
}

var (
	bsdExtendedNameRE = regexp.MustCompile(`^#1\/(\d+)$`)
	gnuExtendedNameRE = regexp.MustCompile(`^\/(\d+)$`)
)

func (r *Reader) parseHeader(hdr *Header) error {
	_, err := io.ReadFull(r.r, r.headerBuffer)
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	err = parseTraditionalHeader(hdr, r.headerBuffer)
	if err != nil {
		return err
	}

	if r.DisableBSDExtensions && r.DisableGnuExtensions {
		return nil
	}

	bsdExtendedNameSize := bsdExtendedNameRE.FindStringSubmatch(hdr.Name)
	gnuExtendedNameOffset := gnuExtendedNameRE.FindStringSubmatch(hdr.Name)

	switch {
	case hdr.Name == gnuExtendedFormatNameTable && !r.DisableGnuExtensions:
		// read the Gnu file name lookup table of
		r.gnuNameBuffer = make([]byte, hdr.Size)

		_, err := r.r.Read(r.gnuNameBuffer)
		if err != nil {
			return fmt.Errorf("read GNU name table")
		}

		nextHeader, err := r.Next()
		if err != nil {
			return err
		}

		// present the next header to the caller
		*hdr = *nextHeader

		return nil
	case len(gnuExtendedNameOffset) == 2 && !r.DisableGnuExtensions:
		// lookup actual file name from the Gnu file name lookup table
		if r.gnuNameBuffer == nil {
			return fmt.Errorf("encountered name reference without prior name table declaration")
		}

		nameOffset, err := strconv.Atoi(gnuExtendedNameOffset[1])
		if err != nil {
			return fmt.Errorf("parse BSD extended name offset: %w", err)
		}

		if nameOffset < 0 {
			return fmt.Errorf("invalid Gnu name table offset: %w", err)
		}

		if nameOffset >= len(r.gnuNameBuffer) {
			return fmt.Errorf( //nolint:stylecheck
				"Gnu name table offset %d out of bounds for table of size %d",
				nameOffset, len(r.gnuNameBuffer))
		}

		// determine end of file name
		terminatorIndex := bytes.Index(r.gnuNameBuffer[nameOffset:], []byte{'\n'})
		if terminatorIndex < 0 {
			terminatorIndex = len(r.gnuNameBuffer) - nameOffset
		}

		hdr.Name = string(r.gnuNameBuffer[nameOffset : nameOffset+terminatorIndex])
	case len(bsdExtendedNameSize) == 2 && !r.DisableBSDExtensions:
		// read actual file name from the beginning of the data section in BSD style
		actualNameSize, err := strconv.Atoi(bsdExtendedNameSize[1])
		if err != nil {
			return fmt.Errorf("parse BSD extended name size: %w", err)
		}

		nameBuffer := make([]byte, actualNameSize)

		_, err = r.r.Read(nameBuffer)
		if err != nil {
			return fmt.Errorf("read BSD extended name: %w", err)
		}

		hdr.Size -= int64(actualNameSize) // correct file size
		hdr.Name = strings.TrimRight(string(nameBuffer), "\x00")
	}

	if !r.DisableGnuExtensions {
		hdr.Name = strings.TrimSuffix(hdr.Name, "/")
	}

	return nil
}

func parseTraditionalHeader(hdr *Header, rawHeader []byte) (err error) {
	if len(rawHeader) != HeaderSize {
		return fmt.Errorf("parsing header requires %d bytes instead of %d",
			HeaderSize, len(rawHeader))
	}

	// check if header ends with the correct byte sequence
	if !bytes.Equal(rawHeader[HeaderSize-len(HeaderTerminator):], []byte(HeaderTerminator)) {
		return fmt.Errorf("unexpected header terminator: %q",
			string(rawHeader[HeaderSize-len(HeaderTerminator):]))
	}

	offset := 0

	hdr.Name = unpackString(rawHeader[offset : offset+nameFieldSize])
	offset += nameFieldSize

	rawModTime, err := unpackUint64(rawHeader[offset : offset+modTimeFieldSize])
	if err != nil {
		return fmt.Errorf("parse mod time: %w", err)
	}

	hdr.ModTime = time.Unix(rawModTime, 0)
	offset += modTimeFieldSize

	uid, err := unpackUint64(rawHeader[offset : offset+uidFieldSize])
	if err != nil {
		return fmt.Errorf("parse uid: %w", err)
	}

	hdr.UID = uid
	offset += uidFieldSize

	gid, err := unpackUint64(rawHeader[offset : offset+gidFieldSize])
	if err != nil {
		return fmt.Errorf("parse uid: %w", err)
	}

	hdr.GID = gid
	offset += gidFieldSize

	hdr.Mode, err = unpackOctal(rawHeader[offset : offset+modeFiledSize])
	if err != nil {
		return fmt.Errorf("parse mode: %w", err)
	}

	offset += modeFiledSize

	hdr.Size, err = unpackUint64(rawHeader[offset : offset+sizeFieldSize])
	if err != nil {
		return fmt.Errorf("parse file size: %w", err)
	}

	return nil
}

func unpackUint64(field []byte) (int64, error) {
	fieldString := extractFromByteField(field)
	if fieldString == "" {
		return 0, nil // Gnu file name lookup tables leave the field empty
	}

	return strconv.ParseInt(fieldString, 10, 64)
}

func unpackOctal(field []byte) (uint32, error) {
	octalStr := strings.TrimPrefix(extractFromByteField(field), "100")
	if octalStr == "" {
		return 0, nil // Gnu file name lookup tables leave the field empty
	}

	i, err := strconv.ParseUint(octalStr, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("parse int: %w", err)
	}

	return uint32(i), nil
}

func unpackString(field []byte) string {
	return extractFromByteField(field)
}

func extractFromByteField(field []byte) string {
	return strings.TrimRight(string(field), " ")
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}

	return 0
}
