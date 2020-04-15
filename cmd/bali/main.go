package main

import (
	"fmt"
	"os"

	"github.com/balibuild/bali/utilities"
)

// global mode
var (
	IsDebugMode bool = false
	IsForceMode bool = false //Overwrite configuration file
)

// version info
var (
	VERSION     = "1.0"
	BUILDTIME   string
	BUILDCOMMIT string
	BUILDBRANCH string
	GOVERSION   string
)

func version() {
	fmt.Fprint(os.Stdout, "Bali - Golang Minimalist build and packaging tool\nversion:       ", VERSION, "\n",
		"build branch:  ", BUILDBRANCH, "\n",
		"build commit:  ", BUILDCOMMIT, "\n",
		"build time:    ", BUILDTIME, "\n",
		"go version:    ", GOVERSION, "\n")
}

func usage() {
	fmt.Fprintf(os.Stdout, `Bali - Golang Minimalist build and packaging tool
usage: %s <option> args ...
  -h|--help        Show usage text and quit
  -v|--version     Show version number and quit
  -V|--verbose     Make the operation more talkative
  -F|--force       Turn on force mode. eg: Overwrite configuration file
  -c|--cwd         Build dir. (Position 0, default cwd)
  -a|--arch        Build arch: amd64 386 arm arm64
  -t|--target      Build target: windows linux darwin ...
  -o|--out         Build output dir. default '$CWD/build'
  -z|--zip         Create archive file after successful build
  -p|--pack        After successful build, create installation package (UNIX STGZ. Windows others)
  -d|--dist        STGZ/TarGz package distribution directory

`, os.Args[0])
}

// DbgPrint todo
func DbgPrint(format string, a ...interface{}) {
	if IsDebugMode {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(utilities.StrCat("\x1b[33m* ", ss, "\x1b[0m\n"))
	}
}

type baliOptions struct {
}

// Invoke argv Receiver
func (bo *baliOptions) Invoke(val int, oa, raw string) error {

	return nil
}

func (bo *baliOptions) parse() error {

	return nil
}

func main() {

}
