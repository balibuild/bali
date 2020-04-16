package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	rel, err := filepath.Rel("/tmp/bali/bali", "/tmp/bali/baligo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "link error %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%s\n", rel)
	if len(os.Args) > 1 {
		source, err := os.Readlink(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "readlink error %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "readlink: [%s]\n", source)
	}
}
