package pecoff

import (
	"strconv"

	"github.com/balibuild/bali/utilities"
)

// Version Int version
type Version struct {
	Major int `json:"Major,omitempty"`
	Minor int `json:"Minor,omitempty"`
	Patch int `json:"Patch,omitempty"`
	Build int `json:"Build,omitempty"`
}

func (ver *Version) fillString(sv string) error {
	svv := utilities.StrSplitSkipEmpty(sv, '.', 4)
	if len(svv) == 0 {
		ver.Patch = 1
		return nil
	}
	var err error
	if len(svv) > 3 {
		ver.Build, err = strconv.Atoi(svv[3])
	}
	if len(svv) > 2 {
		ver.Patch, err = strconv.Atoi(svv[2])
	}
	if len(svv) > 1 {
		ver.Minor, err = strconv.Atoi(svv[2])
	}
	ver.Major, err = strconv.Atoi(svv[2])
	return err
}

// VersionInfo todo
type VersionInfo struct {
	Comment          string `json:"Comment,omitempty"`
	Company          string `json:"Company,omitempty"`
	Description      string `json:"Description,omitempty"`
	FileVersion      string `json:"FileVersion,omitempty"`
	InternalName     string `json:"InternalName,omitempty"`
	LegalCopyright   string `json:"LegalCopyright,omitempty"`
	LegalTrademarks  string `json:"LegalTrademarks,omitempty"`
	OriginalFilename string `json:"OriginalFilename,omitempty"`
	PrivateBuild     string `json:"PrivateBuild,omitempty"`
	ProductName      string `json:"ProductName,omitempty"`
	ProductVersion   string `json:"ProductVersion,omitempty"`
	SpecialBuild     string `json:"SpecialBuild,omitempty"`
	Translation      int    `json:"Translation,omitempty"`
	Charset          int    `json:"Charset,omitempty"`
	fileVersion      Version
	productVersion   Version
}

// Filling todo
func (vi *VersionInfo) Filling() error {
	if err := vi.fileVersion.fillString(vi.FileVersion); err != nil {
		return utilities.ErrorCat("Fill FileVersion: ", err.Error())
	}
	if err := vi.productVersion.fillString(vi.ProductVersion); err != nil {
		return utilities.ErrorCat("Fill ProductVersion: ", err.Error())
	}
	return nil
}
