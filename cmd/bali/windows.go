// +build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// bali init on windows

// const
const (
	EnableVirtualTerminalProcessingMode = 0x4
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func init() {
	var mode uint32
	// becasue we print message to stderr
	h := os.Stderr.Fd()
	if r, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); r != 0 {
		_, _, _ = procSetConsoleMode.Call(h, uintptr(mode|EnableVirtualTerminalProcessingMode))
	}
}
