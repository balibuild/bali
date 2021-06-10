package main

import (
	"fmt"
	"os"

	"github.com/balibuild/bali/inno"
)

func main() {
	fmt.Fprintf(os.Stderr, "Path:    %s\nVersion: %s\n", inno.InnoExePath(), inno.InnoVersion())
}
