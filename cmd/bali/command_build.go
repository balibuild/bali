package main

import (
	"context"
	"strings"

	"github.com/balibuild/bali/v3/pkg/barrow"
)

type BuildCommand struct {
	Target      string   `name:"target" short:"T" help:"Target OS for which the code is compiled" default:"${target}"`       // windows/darwin
	Arch        string   `name:"arch" short:"A" help:"Target architecture for which the code is compiled" default:"${arch}"` // amd64/arm64 ...
	Release     string   `name:"release" help:"Specifies the rpm package tag version"`                                       // --release $TASK_ID
	Destination string   `name:"destination" short:"D" help:"Specify the package save destination" default:"out"`
	Pack        []string `name:"pack" help:"Packaged in a specific format. supported: zip, tar, sh, rpm, deb, apk, arch"`
	Compression string   `name:"compression" help:"Specifies the compression method"`
}

func (c *BuildCommand) Run(g *Globals) error {
	b := barrow.BarrowCtx{
		CWD:         g.M,
		Out:         g.B,
		Target:      c.Target,
		Arch:        c.Arch,
		Release:     c.Release,
		Destination: c.Destination,
		Pack:        c.Pack,
		Compression: strings.ToLower(c.Compression),
		Verbose:     g.Verbose,
	}
	if err := b.Initialize(context.Background()); err != nil {
		return err
	}
	return b.Run(context.Background())
}
