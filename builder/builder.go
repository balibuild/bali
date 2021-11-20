package builder

import (
	"encoding/json"
	"os"

	"github.com/balibuild/bali/v2/base"
	"github.com/balibuild/bali/v2/goversioninfo"
)

// Builder build syso file
type Builder struct {
	vi goversioninfo.VersionInfo
}

// ParseJSON todo
func (b *Builder) ParseJSON(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	if err := json.NewDecoder(fd).Decode(&b.vi); err != nil {
		return err
	}
	return nil
}

// AddIcon add icon to resources
func (b *Builder) AddIcon(src string) error {
	if !base.PathExists(src) {
		return base.ErrorCat("icon: ", src, " not found")
	}
	b.vi.IconPath = src
	return nil
}

// AddManifest todo
func (b *Builder) AddManifest(src string) error {
	if !base.PathExists(src) {
		return base.ErrorCat("manifest: ", src, " not found")
	}
	b.vi.ManifestPath = src
	return nil
}

// FillVersion todo
func (b *Builder) FillVersion(filever, prover string) {
	if len(b.vi.StringFileInfo.FileVersion) == 0 && len(filever) != 0 {
		b.vi.StringFileInfo.FileVersion = filever
	}
	if b.vi.FixedFileInfo.FileVersion.IsZero() {
		_ = b.vi.FixedFileInfo.FileVersion.Fillling(b.vi.StringFileInfo.FileVersion)
	}
	if len(b.vi.StringFileInfo.ProductVersion) == 0 && len(prover) != 0 {
		b.vi.StringFileInfo.ProductVersion = prover
	}
	if b.vi.FixedFileInfo.ProductVersion.IsZero() {
		_ = b.vi.FixedFileInfo.ProductVersion.Fillling(b.vi.StringFileInfo.ProductVersion)
	}
}

// UpdateName todo
func (b *Builder) UpdateName(fileName, productName, desc string) {
	if len(b.vi.StringFileInfo.ProductName) == 0 && len(productName) != 0 {
		b.vi.StringFileInfo.ProductName = productName
	}
	if len(b.vi.StringFileInfo.InternalName) == 0 && len(fileName) != 0 {
		b.vi.StringFileInfo.InternalName = fileName
	}
	if len(b.vi.StringFileInfo.FileDescription) == 0 && len(desc) != 0 {
		b.vi.StringFileInfo.FileDescription = desc
	}
}

// MakeSysoPath todo
func MakeSysoPath(outdir, arch string) string {
	return base.StrCat(outdir, string(os.PathSeparator), "windows_", arch, ".syso")
}

// WriteSyso todo
func (b *Builder) WriteSyso(fileout, arch string) error {
	if len(b.vi.FixedFileInfo.FileFlagsMask) == 0 {
		b.vi.FixedFileInfo.FileFlagsMask = "3f"
	}
	if len(b.vi.FixedFileInfo.FileFlags) == 0 {
		b.vi.FixedFileInfo.FileFlagsMask = "00"
	}
	if len(b.vi.FixedFileInfo.FileOS) == 0 {
		// HEX
		b.vi.FixedFileInfo.FileOS = "40004"
	}
	if len(b.vi.FixedFileInfo.FileType) == 0 {
		b.vi.FixedFileInfo.FileType = "01"
	}
	if len(b.vi.FixedFileInfo.FileSubType) == 0 {
		b.vi.FixedFileInfo.FileSubType = "00"
	}
	if b.vi.VarFileInfo.Translation.LangID == 0 {
		b.vi.VarFileInfo.Translation.LangID = goversioninfo.LngUSEnglish
	}
	if b.vi.VarFileInfo.Translation.CharsetID == 0 {
		b.vi.VarFileInfo.Translation.CharsetID = goversioninfo.CsUnicode
	}
	b.vi.Build()
	b.vi.Walk()
	if err := b.vi.WriteSyso(fileout, arch); err != nil {
		return base.ErrorCat("Error writing syso: ", err.Error())
	}
	return nil
}
