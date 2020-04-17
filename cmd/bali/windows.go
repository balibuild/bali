// +build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// bali init on windows

const (
	cENABLEVIRTUALTERMINALPROCESSING = 0x4
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// enableColorsStdout
func init() {
	var mode uint32
	h := os.Stdout.Fd()
	if r, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); r != 0 {
		procSetConsoleMode.Call(h, uintptr(mode|cENABLEVIRTUALTERMINALPROCESSING))
	}
}
