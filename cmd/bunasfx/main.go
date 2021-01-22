package main

import (
	"fmt"
	"os"
)

func openExecutable() (*os.File, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	fd, err := os.Open(exe)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func main() {
	fd, err := openExecutable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable open executable: %v\n", err)
		os.Exit(1)
	}
	defer fd.Close()
	st, err := fd.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable stat %v\n", err)
		os.Exit(1)
	}
	offset, err := overlayOffset(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable detect overlay offset %v\n", err)
		os.Exit(1)
	}
	if offset == st.Size() {
		fmt.Fprintf(os.Stderr, "executable file does not seem to contain enough additional data\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%d %d\n", offset, st.Size())
}
