package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s pefile\n", os.Args[0])
		os.Exit(1)
	}
	a := NewAssets(os.Args[1])
	if err := a.Parse(); err != nil {
		fmt.Fprintf(os.Stderr, "unable parse pefile: %v\n", err)
		os.Exit(1)
	}
	baseName := strings.TrimSuffix(filepath.Base(os.Args[1]), ".exe") + ".zip"
	if err := a.Write(baseName); err != nil {
		fmt.Fprintf(os.Stderr, "unable write file: %v\n", err)
		os.Exit(1)
	}
}
