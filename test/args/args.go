package main

import (
	"encoding/json"
	"os"
	"runtime"

	"github.com/alecthomas/kong"
)

type VersionFlag bool

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	app.Exit(0)
	return nil
}

type Globals struct {
	M       string      `name:"module" short:"M" help:"Explicitly specify a module directory" default:"." type:"path"`
	B       string      `name:"build" short:"B" help:"Explicitly specify a build directory" default:"build" type:"path"`
	Verbose bool        `name:"verbose" short:"V" help:"Make the operation more talkative"`
	Version VersionFlag `name:"version" short:"v" help:"Print version information and quit"`
}

type CleanCommand struct {
}

func (c *CleanCommand) Run(g *Globals) error {

	return nil
}

type BuildCommand struct {
	Target string `name:"target" short:"T" help:"Target OS for which the code is compiled" default:"${target}"`       // windows/darwin
	Arch   string `name:"arch" short:"A" help:"Target architecture for which the code is compiled" default:"${arch}"` // amd64/arm64 ...
}

func (c *BuildCommand) Run(g *Globals) error {
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("  ", "")
	enc.Encode(g)
	enc.Encode(c)
	return nil
}

type App struct {
	Globals
	Build BuildCommand `cmd:"build" help:"Compile the current module (default)" default:"withargs"`
	Clean CleanCommand `cmd:"clean" help:"Remove generated artifacts"`
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
	ctx.FatalIfErrorf(err)
}
