// Package goversioninfo creates a syso file which contains Microsoft Version Information and an optional icon.
package goversioninfo

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/balibuild/bali/v3/module/rsrc/binutil"
	"github.com/balibuild/bali/v3/module/rsrc/coff"
)

// VersionInfo data container
type VersionInfo struct {
	Icon           string         `toml:"icon,omitempty" json:"icon,omitempty"`
	Manifest       string         `toml:"manifest,omitempty" json:"manifest,omitempty"` // path or content
	FixedFileInfo  FixedFileInfo  `toml:"FixedFileInfo,omitempty" json:"FixedFileInfo,omitempty"`
	StringFileInfo StringFileInfo `toml:"StringFileInfo,omitempty" json:"StringFileInfo,omitempty"`
	VarFileInfo    VarFileInfo    `toml:"VarFileInfo,omitempty" json:"VarFileInfo,omitempty"`
	Timestamp      bool           `toml:"Timestamp,omitempty" json:"Timestamp,omitempty"`
	structure      VSVersionInfo
	buffer         bytes.Buffer
}

// Translation with langid and charsetid.
type Translation struct {
	LangID    LangID    `toml:"LangID,omitempty" json:"LangID,omitempty"`
	CharsetID CharsetID `toml:"CharsetID,omitempty" json:"CharsetID,omitempty"`
}

// FileVersion with 3 parts.
type FileVersion struct {
	Major int `toml:"Major,omitempty" json:"Major,omitempty"`
	Minor int `toml:"Minor,omitempty" json:"Minor,omitempty"`
	Patch int `toml:"Patch,omitempty" json:"Patch,omitempty"`
	Build int `toml:"Build,omitempty" json:"Build,omitempty"`
}

// Overwrite version
func (fv *FileVersion) Overwrite(ver string) error {
	if len(ver) == 0 {
		return nil
	}
	vss := strings.Split(ver, ".")
	if len(vss) > 3 && fv.Build == 0 {
		fv.Build, _ = strconv.Atoi(vss[3])
	}
	if len(vss) > 2 && fv.Patch == 0 {
		fv.Patch, _ = strconv.Atoi(vss[2])
	}
	if len(vss) > 1 && fv.Minor == 0 {
		fv.Minor, _ = strconv.Atoi(vss[1])
	}
	var err error
	fv.Major, err = strconv.Atoi(vss[0])
	return err
}

// FixedFileInfo contains file characteristics - leave most of them at the defaults.
type FixedFileInfo struct {
	FileVersion    FileVersion `toml:"FileVersion,omitempty" json:"FileVersion,omitempty"`
	ProductVersion FileVersion `toml:"ProductVersion,omitempty" json:"ProductVersion,omitempty"`
	FileFlagsMask  string      `toml:"FileFlagsMask,omitempty" json:"FileFlagsMask,omitempty"`
	FileFlags      string      `toml:"FileFlags,omitempty" json:"FileFlags,omitempty"`
	FileOS         string      `toml:"FileOS,omitempty" json:"FileOS,omitempty"`
	FileType       string      `toml:"FileType,omitempty" json:"FileType,omitempty"`
	FileSubType    string      `toml:"FileSubType,omitempty" json:"FileSubType,omitempty"`
}

// VarFileInfo is the translation container.
type VarFileInfo struct {
	Translation `toml:"Translation,omitempty" json:"Translation,omitempty"`
}

// StringFileInfo is what you want to change.
type StringFileInfo struct {
	Comments         string `toml:"Comments,omitempty" json:"Comments,omitempty"`
	CompanyName      string `toml:"CompanyName,omitempty" json:"CompanyName,omitempty"`
	FileDescription  string `toml:"FileDescription,omitempty" json:"FileDescription,omitempty"`
	FileVersion      string `toml:"FileVersion,omitempty" json:"FileVersion,omitempty"`
	InternalName     string `toml:"InternalName,omitempty" json:"InternalName,omitempty"`
	LegalCopyright   string `toml:"LegalCopyright,omitempty" json:"LegalCopyright,omitempty"`
	LegalTrademarks  string `toml:"LegalTrademarks,omitempty" json:"LegalTrademarks,omitempty"`
	OriginalFilename string `toml:"OriginalFilename,omitempty" json:"OriginalFilename,omitempty"`
	PrivateBuild     string `toml:"PrivateBuild,omitempty" json:"PrivateBuild,omitempty"`
	ProductName      string `toml:"ProductName,omitempty" json:"ProductName,omitempty"`
	ProductVersion   string `toml:"ProductVersion,omitempty" json:"ProductVersion,omitempty"`
	SpecialBuild     string `toml:"SpecialBuild,omitempty" json:"SpecialBuild,omitempty"`
}

// *****************************************************************************
// Helpers
// *****************************************************************************

// SizedReader is a *bytes.Buffer.
type SizedReader struct {
	*bytes.Buffer
}

// Size returns the length of the buffer.
func (s SizedReader) Size() int64 {
	return int64(s.Len())
}

func str2Uint32(s string) uint32 {
	if s == "" {
		return 0
	}
	u, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		log.Printf("Error parsing %q as uint32: %v", s, err)
		return 0
	}

	return uint32(u)
}

func padString(s string, zeros int) []byte {
	b := make([]byte, 0, len([]rune(s))*2)
	for _, x := range s {
		tt := int32(x)

		b = append(b, byte(tt))
		if tt > 255 {
			tt = tt >> 8
			b = append(b, byte(tt))
		} else {
			b = append(b, byte(0))
		}
	}

	for i := 0; i < zeros; i++ {
		b = append(b, 0x00)
	}

	return b
}

func padBytes(i int) []byte {
	return make([]byte, i)
}

func (f *FileVersion) getVersionHighString() string {
	return fmt.Sprintf("%04x%04x", f.Major, f.Minor)
}

func (f *FileVersion) getVersionLowString() string {
	return fmt.Sprintf("%04x%04x", f.Patch, f.Build)
}

// GetVersionString returns a string representation of the version
func (f *FileVersion) GetVersionString() string {
	return fmt.Sprintf("%d.%d.%d.%d", f.Major, f.Minor, f.Patch, f.Build)
}

func (t Translation) getTranslationString() string {
	return fmt.Sprintf("%04X%04X", t.LangID, t.CharsetID)
}

func (t Translation) getTranslation() string {
	return fmt.Sprintf("%04x%04x", t.CharsetID, t.LangID)
}

// *****************************************************************************
// IO Methods
// *****************************************************************************

// Walk writes the data buffer with hexadecimal data from the structs
func (vi *VersionInfo) Walk() {
	// Create a buffer
	var b bytes.Buffer
	w := binutil.Writer{W: &b}

	// Write to the buffer
	binutil.Walk(vi.structure, func(v reflect.Value, path string) error {
		if binutil.Plain(v.Kind()) {
			w.WriteLE(v.Interface())
		}
		return nil
	})

	vi.buffer = b
}

func (vi *VersionInfo) loadManifest(cwd string, rsrc *coff.Coff, newID chan uint16) error {
	if manifest, ok := strings.CutPrefix(vi.Manifest, "data:"); ok {
		id := <-newID
		rsrc.AddResource(rtManifest, id, strings.NewReader(manifest))
		return nil
	}
	if len(vi.Manifest) == 0 {
		return nil
	}

	fd, err := binutil.SizedOpen(filepath.Join(cwd, vi.Manifest))
	if err != nil {
		return err
	}
	defer fd.Close()

	id := <-newID
	rsrc.AddResource(rtManifest, id, fd)
	return nil
}

// WriteSyso creates a resource file from the version info and optionally an icon.
// arch must be an architecture string accepted by coff.Arch, like "386" or "amd64" waiting support "arm"  and "arm64"
func (vi *VersionInfo) WriteSyso(cwd string, saveTo string, arch string) error {

	// Channel for generating IDs
	newID := make(chan uint16)
	go func() {
		for i := uint16(1); ; i++ {
			newID <- i
		}
	}()

	// Create a new RSRC section
	rsrc := coff.NewRSRC()

	// Set the architecture
	err := rsrc.Arch(arch)
	if err != nil {
		return err
	}

	// ID 16 is for Version Information
	rsrc.AddResource(16, 1, SizedReader{bytes.NewBuffer(vi.buffer.Bytes())})
	if err := vi.loadManifest(cwd, rsrc, newID); err != nil {
		return err
	}
	// If icon is enabled
	if vi.Icon != "" {
		if err := addIcon(rsrc, filepath.Join(cwd, vi.Icon), newID); err != nil {
			return err
		}
	}

	rsrc.Freeze()

	// Write to file
	return writeCoff(rsrc, saveTo)
}

// WriteHex creates a hex file for debugging version info
func (vi *VersionInfo) WriteHex(saveTo string) error {
	return os.WriteFile(saveTo, vi.buffer.Bytes(), 0655)
}

func writeCoff(coff *coff.Coff, saveTo string) error {
	out, err := os.Create(saveTo)
	if err != nil {
		return err
	}
	if err = writeCoffTo(out, coff); err != nil {
		return fmt.Errorf("error writing %q: %v", saveTo, err)
	}
	return nil
}

func writeCoffTo(w io.WriteCloser, coff *coff.Coff) error {
	bw := binutil.Writer{W: w}

	// write the resulting file to disk
	binutil.Walk(coff, func(v reflect.Value, path string) error {
		if binutil.Plain(v.Kind()) {
			bw.WriteLE(v.Interface())
			return nil
		}
		vv, ok := v.Interface().(binutil.SizedReader)
		if ok {
			bw.WriteFromSized(vv)
			return binutil.ErrWalkSkip
		}
		return nil
	})

	err := bw.Err
	if closeErr := w.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	return err
}
