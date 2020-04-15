package main

import (
	"fmt"
	"os"
	"path/filepath"

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
  -w|--workdir     Specify bali running directory. (Position 0, default $PWD)
  -a|--arch        Build arch: amd64 386 arm arm64
  -t|--target      Build target: windows linux darwin ...
  -o|--out         Specify build output directory. default '$PWD/build'
  -d|--dest        Specify the path to save the package
  -z|--zip         Create archive file (UNIX: .tar.gz, Windows: .zip)
  -p|--pack        Create installation package (UNIX: STGZ, Windows: none)

`, os.Args[0])
}

// DbgPrint todo
func DbgPrint(format string, a ...interface{}) {
	if IsDebugMode {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(utilities.StrCat("\x1b[33m* ", ss, "\x1b[0m\n"))
	}
}

// Invoke argv Receiver
func (be *BaliExecutor) Invoke(val int, oa, raw string) error {
	switch val {
	case 'h':
		usage()
		os.Exit(0)
	case 'v':
		version()
		os.Exit(0)
	case 'c':
		cwd, err := filepath.Abs(oa)
		if err != nil {
			return err
		}
		be.workdir = cwd
	case 'a':
		be.arch = oa
	case 't': // target
		be.target = oa
	case 'V':
		IsDebugMode = true
	case 'f': // --force
		IsForceMode = true
	case 'o': // --out
		out, err := filepath.Abs(oa)
		if err != nil {
			return err
		}
		be.out = out
	case 'd':
		dest, err := filepath.Abs(oa)
		if err != nil {
			return err
		}
		be.destination = dest
	case 'z': // --zip
		be.makezip = true
	case 'p': // --pack
		be.makepack = true
	default:
	}
	return nil
}

func (be *BaliExecutor) parse() error {
	var ae utilities.ArgvParser
	ae.Add("help", utilities.NOARG, 'h')
	ae.Add("version", utilities.NOARG, 'v')
	ae.Add("verbose", utilities.NOARG, 'V')
	ae.Add("force", utilities.NOARG, 'f')
	ae.Add("workdir", utilities.REQUIRED, 'w')
	ae.Add("arch", utilities.REQUIRED, 'a')
	ae.Add("target", utilities.REQUIRED, 't')
	ae.Add("out", utilities.REQUIRED, 'o')
	ae.Add("dest", utilities.REQUIRED, 'd')
	ae.Add("zip", utilities.NOARG, 'z')
	ae.Add("pack", utilities.NOARG, 'p')
	if err := ae.Execute(os.Args, be); err != nil {
		return err
	}
	if len(be.workdir) == 0 {
		if len(ae.Unresolved()) > 0 {
			cwd, err := filepath.Abs(ae.Unresolved()[0])
			if err != nil {
				return err
			}
			be.workdir = cwd
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			be.workdir = cwd
		}
	}
	// set build out dir
	if len(be.out) == 0 {
		be.out = filepath.Join(be.workdir, "build")
	}
	if len(be.destination) == 0 {
		be.destination = be.workdir
	}
	return nil
}

func main() {
	var be BaliExecutor
	if err := be.parse(); err != nil {
		fmt.Fprintf(os.Stderr, "bali: parse args: \x1b[31m%v\x1b[0m\n", err)
		os.Exit(1)
	}
}
