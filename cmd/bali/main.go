package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/balibuild/bali/pack"
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
	fmt.Fprint(os.Stdout, "Bali - Minimalist Golang build and packaging tool\nversion:       ", VERSION, "\n",
		"build branch:  ", BUILDBRANCH, "\n",
		"build commit:  ", BUILDCOMMIT, "\n",
		"build time:    ", BUILDTIME, "\n",
		"go version:    ", GOVERSION, "\n")
}

func usage() {
	fmt.Fprintf(os.Stdout, `Bali - Minimalist Golang build and packaging tool
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
  -A|--method      Zip compress method: zstd xz bzip2 deflate(default)
  --cleanup        Cleanup build directory
  --no-rename      Disable file renaming (STGZ installation package, default: OFF)

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
func (be *Executor) Invoke(val int, oa, raw string) error {
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
	case 'A':
		oa = strings.ToLower(oa)
		switch oa {
		case "deflate":
			be.zipmethod = pack.Deflate
		case "bzip2":
			be.zipmethod = pack.BZIP2
		case "zstd":
			be.zipmethod = pack.ZSTD
		case "xz":
			be.zipmethod = pack.XZ
		}
	case 'p': // --pack
		be.makepack = true
	case 1001:
		be.cleanup = true
	case 1002:
		be.norename = true
	default:
	}
	return nil
}

// ParseArgv parse argv
func (be *Executor) ParseArgv() error {
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
	ae.Add("method", utilities.REQUIRED, 'A')
	ae.Add("pack", utilities.NOARG, 'p')
	ae.Add("cleanup", utilities.NOARG, 1001)
	ae.Add("no-rename", utilities.NOARG, 1002)
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
	var be Executor
	if err := be.ParseArgv(); err != nil {
		fmt.Fprintf(os.Stderr, "bali: parse args: \x1b[31m%v\x1b[0m\n", err)
		os.Exit(1)
	}
	if be.cleanup {
		if err := os.RemoveAll(be.out); err != nil {
			fmt.Fprintf(os.Stderr, "bali: cleanup %s: \x1b[31m%v\x1b[0m\n", be.out, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "\x1b[32mbali: cleanup %s success\x1b[0m\n", be.out)
		os.Exit(0)
	}
	if err := be.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "bali: initialize: \x1b[31m%v\x1b[0m\n", err)
		os.Exit(1)
	}
	if err := be.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "bali: build: \x1b[31m%v\x1b[0m\n", err)
		os.Exit(1)
	}
	if be.makezip {
		if err := be.Compress(); err != nil {
			fmt.Fprintf(os.Stderr, "bali: compress: \x1b[31m%v\x1b[0m\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "bali create archive \x1b[32m%s\x1b[0m success\n", be.bm.Name)
	}
	if be.makepack {
		if err := be.Pack(); err != nil {
			fmt.Fprintf(os.Stderr, "bali: make pack: \x1b[31m%v\x1b[0m\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "bali create package \x1b[32m%s\x1b[0m success\n", be.bm.Name)
	}
}
