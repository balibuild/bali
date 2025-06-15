package barrow

import (
	"os"
	"path/filepath"

	"github.com/balibuild/bali/v3/module/goversioninfo"
)

func (b *BarrowCtx) makeResources(e *Crate, saveTo string) error {
	var vi goversioninfo.VersionInfo
	if err := LoadMetadata(filepath.Join(e.cwd, "winres.toml"), &vi); err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(vi.StringFileInfo.FileVersion) == 0 {
		vi.StringFileInfo.FileVersion = e.Version
	}
	vi.FixedFileInfo.FileVersion.Overwrite(e.Version)
	if len(vi.StringFileInfo.ProductVersion) == 0 {
		vi.StringFileInfo.ProductVersion = e.Version
	}
	vi.FixedFileInfo.ProductVersion.Overwrite(e.Version)
	if len(vi.StringFileInfo.ProductName) == 0 {
		vi.StringFileInfo.ProductName = e.Name
	}
	if len(vi.StringFileInfo.InternalName) == 0 {
		vi.StringFileInfo.InternalName = b.binaryName(e.Name)
	}
	if len(vi.StringFileInfo.FileDescription) == 0 {
		vi.StringFileInfo.FileDescription = e.Description
	}
	if len(vi.FixedFileInfo.FileFlagsMask) == 0 {
		vi.FixedFileInfo.FileFlagsMask = "3f"
	}
	if len(vi.FixedFileInfo.FileFlags) == 0 {
		vi.FixedFileInfo.FileFlagsMask = "00"
	}
	if len(vi.FixedFileInfo.FileOS) == 0 {
		// HEX
		vi.FixedFileInfo.FileOS = "40004"
	}
	if len(vi.FixedFileInfo.FileType) == 0 {
		vi.FixedFileInfo.FileType = "01"
	}
	if len(vi.FixedFileInfo.FileSubType) == 0 {
		vi.FixedFileInfo.FileSubType = "00"
	}
	if vi.VarFileInfo.LangID == 0 {
		vi.VarFileInfo.LangID = goversioninfo.LngUSEnglish
	}
	if vi.VarFileInfo.CharsetID == 0 {
		vi.VarFileInfo.CharsetID = goversioninfo.CsUnicode
	}
	vi.Build()
	vi.Walk()
	return vi.WriteSyso(e.cwd, saveTo, b.Arch)
}
