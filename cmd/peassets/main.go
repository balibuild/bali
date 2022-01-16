package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/balibuild/bali/v2/base"
)

// global mode
var (
	IsDebugMode bool = false
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
	const template = `peassets - PE executable program depends on aggregation tool
Version:     %s
Branch:      %s
Commit:      %s
Build Time:  %s
Go Version:  %s

`
	const tagTemplate = `peassets - PE executable program depends on aggregation tool
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
	fmt.Fprintf(os.Stdout, `peassets - PE executable program depends on aggregation tool
usage: %s <option> args ...
  -h|--help          Show usage text and quit
  -v|--version       Show version number and quit
  -V|--verbose       Make the operation more talkative
  -d|--dest          Specify the path to save the package

`, os.Args[0])
}

// DbgPrint todo
func DbgPrint(format string, a ...interface{}) {
	if IsDebugMode {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(base.StrCat("\x1b[33m* ", ss, "\x1b[0m\n"))
	}
}

type options struct {
	destination string
	files       []string
}

// Invoke argv Receiver
func (opt *options) Invoke(val int, oa, raw string) error {
	switch val {
	case 'h':
		usage()
		os.Exit(0)
	case 'v':
		version()
		os.Exit(0)
	case 'V':
		IsDebugMode = true
	case 'd':
		dest, err := filepath.Abs(oa)
		if err != nil {
			return err
		}
		opt.destination = dest
	default:
	}
	return nil
}

// ParseArgv parse argv
func (opt *options) ParseArgv() error {
	var pa base.ParseArgs
	pa.Add("help", base.NOARG, 'h')
	pa.Add("version", base.NOARG, 'v')
	pa.Add("verbose", base.NOARG, 'V')
	pa.Add("dest", base.REQUIRED, 'd')
	if err := pa.Execute(os.Args, opt); err != nil {
		return err
	}
	unresolvedArgs := pa.Unresolved()
	if len(unresolvedArgs) == 0 {
		return errors.New("no input files")
	}
	files := make([]string, 0, len(unresolvedArgs))
	unrecorded := func(p string) bool {
		for _, file := range files {
			if strings.EqualFold(file, p) {
				return false
			}
		}
		return true
	}
	for _, a := range unresolvedArgs {
		p, err := filepath.Abs(a)
		if err != nil {
			fmt.Fprintf(os.Stderr, "peassets unresolved file '%s' error: %v\n", a, err)
			continue
		}
		if unrecorded(p) {
			files = append(files, p)
		}
	}
	if len(files) == 0 {
		return errors.New("no input files")
	}
	if len(opt.destination) == 0 {
		opt.destination = strings.TrimSuffix(filepath.Base(files[0]), ".exe") + ".zip"
	}
	opt.files = files
	return nil
}

func main() {
	opt := &options{}
	if err := opt.ParseArgv(); err != nil {
		fmt.Fprintf(os.Stderr, "peassets: %v\n", err)
		os.Exit(1)
	}
	a := NewAssets(opt.files)
	if err := a.Parse(); err != nil {
		fmt.Fprintf(os.Stderr, "peassets parse error: %v\n", err)
		os.Exit(1)
	}
	if err := a.Write(opt.destination); err != nil {
		fmt.Fprintf(os.Stderr, "peassets write zip: %v\n", err)
		os.Exit(1)
	}
}
