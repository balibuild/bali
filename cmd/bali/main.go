package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/balibuild/bali/v2/base"
	"github.com/balibuild/bali/v2/pack"
)

// global mode
var (
	IsDebugMode bool = false
	IsForceMode bool = false //Overwrite configuration file
)

// version info
var (
	VERSION             = "2.1.1"
	BUILDTIME    string = "NONE"
	BUILDCOMMIT  string = "NONE"
	BUILDBRANCH  string = "NONE"
	BUILDREFNAME string = "NONE"
	GOVERSION    string
)

func init() {
	if len(GOVERSION) == 0 {
		GOVERSION = fmt.Sprintf("%s %s/%s", strings.Replace(runtime.Version(), "go", "", 1), runtime.GOOS, runtime.GOARCH)
	}
}

func version() {
	const template = `Bali - Minimalist Golang build and packaging tool
Version:     %s
Branch:      %s
Commit:      %s
Build Time:  %s
Go Version:  %s

`
	const tagTemplate = `Bali - Minimalist Golang build and packaging tool
Version:     %s
Release:     %s
Commit:      %s
Build Time:  %s
Go Version:  %s

`
	if len(BUILDBRANCH) != 0 {
		fmt.Fprintf(os.Stdout, template, VERSION, BUILDBRANCH, BUILDCOMMIT, BUILDTIME, GOVERSION)
		return
	}
	fmt.Fprintf(os.Stdout, tagTemplate, VERSION, strings.TrimPrefix(BUILDREFNAME, "refs/tags/"), BUILDCOMMIT, BUILDTIME, GOVERSION)
}

func usage() {
	fmt.Fprintf(os.Stdout, `Bali - Minimalist Golang build and packaging tool
usage: %s <option> args ...
  -h|--help          Show usage text and quit
  -v|--version       Show version number and quit
  -V|--verbose       Make the operation more talkative
  -F|--force         Turn on force mode. eg: Overwrite configuration file
  -w|--workdir       Specify bali running directory. (Position 0, default $PWD)
  -a|--arch          Build arch: amd64 386 arm arm64
  -t|--target        Build target: windows linux darwin ...
  -o|--out           Specify build output directory. default '$PWD/build'
  -d|--dest          Specify the path to save the package
  -z|--zip           Create archive file (UNIX: .tar.gz, Windows: .zip)
  -p|--pack          Create installation package (UNIX: STGZ, Windows: none)
  -A|--method        Zip compress method: zstd bzip2 brotli deflate(default)
  --cleanup          Cleanup build directory
  --no-rename        Disable file renaming (STGZ installation package, default: OFF)
  --force-version    Force to specify the version of the current project
  --without-version  Package file name without version   

`, os.Args[0])
}

// DbgPrint todo
func DbgPrint(format string, a ...interface{}) {
	if IsDebugMode {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(base.StrCat("\x1b[33m* ", ss, "\x1b[0m\n"))
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
		case "brotli":
			be.zipmethod = pack.BROTLI
		}
	case 'p': // --pack
		be.makepack = true
	case 1001:
		be.cleanup = true
	case 1002:
		be.norename = true
	case 1003:
		be.forceVerion = oa
	case 1004:
		be.withoutVersion = true
	default:
	}
	return nil
}

// ParseArgv parse argv
func (be *Executor) ParseArgv() error {
	var pa base.ParseArgs
	pa.Add("help", base.NOARG, 'h')
	pa.Add("version", base.NOARG, 'v')
	pa.Add("verbose", base.NOARG, 'V')
	pa.Add("force", base.NOARG, 'f')
	pa.Add("workdir", base.REQUIRED, 'w')
	pa.Add("arch", base.REQUIRED, 'a')
	pa.Add("target", base.REQUIRED, 't')
	pa.Add("out", base.REQUIRED, 'o')
	pa.Add("dest", base.REQUIRED, 'd')
	pa.Add("zip", base.NOARG, 'z')
	pa.Add("method", base.REQUIRED, 'A')
	pa.Add("pack", base.NOARG, 'p')
	pa.Add("cleanup", base.NOARG, 1001)
	pa.Add("no-rename", base.NOARG, 1002)
	pa.Add("force-version", base.REQUIRED, 1003)
	pa.Add("without-version", base.NOARG, 1004)
	be.zipmethod = pack.Deflate
	if err := pa.Execute(os.Args, be); err != nil {
		return err
	}
	if len(be.workdir) == 0 {
		if len(pa.Unresolved()) > 0 {
			cwd, err := filepath.Abs(pa.Unresolved()[0])
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
	}
	if be.makepack {
		if err := be.Pack(); err != nil {
			fmt.Fprintf(os.Stderr, "bali: make pack: \x1b[31m%v\x1b[0m\n", err)
			os.Exit(1)
		}
	}
}
