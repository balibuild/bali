package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"unicode"

	"github.com/alecthomas/kong"
)

// version info
var (
	VERSION                = "3.1.1"
	BUILD_TIME      string = "NONE"
	BUILD_COMMIT    string = "NONE"
	BUILD_BRANCH    string = "NONE"
	BUILD_REFNAME   string = "NONE"
	BUILD_GOVERSION string
)

func init() {
	if len(BUILD_GOVERSION) == 0 {
		BUILD_GOVERSION = fmt.Sprintf("%s %s/%s", strings.Replace(runtime.Version(), "go", "", 1), runtime.GOOS, runtime.GOARCH)
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
	if len(BUILD_BRANCH) != 0 {
		fmt.Fprintf(os.Stdout, template, VERSION, BUILD_BRANCH, BUILD_COMMIT, BUILD_TIME, BUILD_GOVERSION)
		return
	}
	fmt.Fprintf(os.Stdout, tagTemplate, VERSION, strings.TrimPrefix(BUILD_REFNAME, "refs/tags/"), BUILD_COMMIT, BUILD_TIME, BUILD_GOVERSION)
}

type VersionFlag bool

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	version()
	app.Exit(0)
	return nil
}

type Globals struct {
	M       string      `name:"module" short:"M" help:"Explicitly specify a module directory" default:"." type:"path"`
	B       string      `name:"build" short:"B" help:"Explicitly specify a build directory" default:"build" type:"path"`
	Verbose bool        `name:"verbose" short:"V" help:"Make the operation more talkative"`
	Version VersionFlag `name:"version" short:"v" help:"Print version information and quit"`
}

func (g *Globals) DbgPrint(format string, a ...any) {
	if !g.Verbose {
		return
	}
	message := fmt.Sprintf(format, a...)
	message = strings.TrimRightFunc(message, unicode.IsSpace)
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		fmt.Fprintf(os.Stderr, "\x1b[33m* %s\x1b[0m\n", line)
	}
}

type App struct {
	Globals
	Build  BuildCommand  `cmd:"build" help:"Compile the current module (default)" default:"withargs"`
	Update UpdateCommand `cmd:"update" help:"Update dependencies as recorded in the go.mod"`
	Clean  CleanCommand  `cmd:"clean" help:"Remove generated artifacts"`
}

func main() {
	app := App{}

	ctx := kong.Parse(&app,
		kong.Name("bali"),
		kong.Description("Bali - Minimalist Golang build and packaging tool"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"target": runtime.GOOS,
			"arch":   runtime.GOARCH,
		})
	err := ctx.Run(&app.Globals)
	if err != nil {
		os.Exit(1)
	}
}
