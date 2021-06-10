// +build windows

package inno

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func resolveInstalledInno() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Inno Setup 6_is1`, registry.QUERY_VALUE)
	if err != nil {
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Inno Setup 6_is1`, registry.QUERY_VALUE)
	}
	if err != nil {
		return "", err
	}
	defer k.Close()
	installLocation, _, err := k.GetStringValue("InstallLocation")
	if err != nil {
		return "", nil
	}
	innoExe := filepath.Join(installLocation, "ISCC.exe")
	if _, err := os.Stat(innoExe); err != nil {
		return "", err
	}
	innoVersion, _, _ = k.GetStringValue("DisplayVersion")
	return innoExe, nil
}

func init() {
	innoExe, err := exec.LookPath("ISCC.exe")
	if err == nil {
		innoExePath = innoExe
		return
	}
	innoExe, err = resolveInstalledInno()
	if err == nil {
		innoExePath = innoExe
		return
	}
	if innoLocation := os.Getenv("BALI_INNO_SETUP_LOCATION"); len(innoLocation) != 0 {
		innoExe := filepath.Join(innoLocation, "ISCC.exe")
		if _, err := os.Stat(innoExe); err == nil {
			innoExePath = innoExe
		}
	}
}
