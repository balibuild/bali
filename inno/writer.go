package inno

import (
	"fmt"
	"os"
	"path/filepath"
)

type Writer struct {
	fd *os.File
}

type Version struct {
	AppId           string `toml:"AppId,omitempty"`
	AppName         string `toml:"AppName,omitempty"`
	AppVersion      string `toml:"AppVersion,omitempty"`
	AppVerName      string `toml:"AppVerName,omitempty"`
	AppPublisher    string `toml:"AppPublisher,omitempty"`
	AppPublisherURL string `toml:"AppPublisherURL,omitempty"`
	AppSupportURL   string `toml:"AppSupportURL,omitempty"`
	LicenseFile     string `toml:"LicenseFile,omitempty"`
}

func NewWriter(buidlDir string) (*Writer, error) {
	p := filepath.Join(buidlDir, "build.iss")
	fd, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	_, _ = fd.Write([]byte{0xEF, 0xBB, 0xBF}) // Write BOM
	fmt.Fprintf(fd, `;; Bali generated file: DO NOT EDIT
;; Please Inno Setup >= 6.2.0
#if Ver < EncodeVer(6,1,0,0)
  #error This script requires Inno Setup 6 or later
#endif
`)
	return &Writer{fd: fd}, nil
}

func (w *Writer) AddTarget() error {

	return nil
}
