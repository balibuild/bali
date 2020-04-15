package builder

import (
	"encoding/json"
	"os"

	"github.com/balibuild/bali/goversioninfo"
	"github.com/balibuild/bali/utilities"
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
	if !utilities.PathExists(src) {
		return utilities.ErrorCat("icon: ", src, " not found")
	}
	b.vi.IconPath = src
	return nil
}

// AddManifest todo
func (b *Builder) AddManifest(src string) error {
	if !utilities.PathExists(src) {
		return utilities.ErrorCat("manifest: ", src, " not found")
	}
	b.vi.ManifestPath = src
	return nil
}

// Version todo
func (b *Builder) Version(filever, prover string) {
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

// Name todo
func (b *Builder) Name(fileName, productName string) {
	if len(b.vi.StringFileInfo.ProductName) == 0 && len(productName) != 0 {
		b.vi.StringFileInfo.ProductName = productName
	}
	if len(b.vi.StringFileInfo.InternalName) == 0 && len(fileName) != 0 {
		b.vi.StringFileInfo.InternalName = fileName
	}
}

// WriteSyso todo
func (b *Builder) WriteSyso(outdir, arch string) error {
	b.vi.Build()
	b.vi.Walk()
	fileout := utilities.StrCat(outdir, "/windows_", arch, ".syso")
	if err := b.vi.WriteSyso(fileout, arch); err != nil {
		return utilities.ErrorCat("Error writing syso: ", err.Error())
	}
	return nil
}
