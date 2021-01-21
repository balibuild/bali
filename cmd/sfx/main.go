package main

import (
	"fmt"
	"os"

	"github.com/fcharlie/buna/debug/pe"
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable resolve executable: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Executable: %s\n", exe)
	fd, err := pe.Open(exe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "not pe file executable: %v\n", err)
		os.Exit(1)
	}
	defer fd.Close()
}
