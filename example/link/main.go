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
}
