package main

import (
	"context"

	"github.com/balibuild/bali/v3/pkg/barrow"
)

type BuildCommand struct {
	Target  string `name:"target" short:"T" help:"Target OS for which the code is compiled" default:"${target}"`       // windows/darwin
	Arch    string `name:"arch" short:"A" help:"Target architecture for which the code is compiled" default:"${arch}"` // amd64/arm64 ...
	Release string `name:"release" help:"Specifies the rpm package tag version"`                                       // --release $TASK_ID
	Pack    string `name:"pack" help:"Pack in the specified format after the build is complete"`
}

func (c *BuildCommand) Run(g *Globals) error {
	b := barrow.BarrowCtx{
		CWD:     g.M,
		Out:     g.B,
		Target:  c.Target,
		Arch:    c.Arch,
		Release: c.Release,
		Pack:    c.Pack,
		Verbose: g.Verbose,
	}
	if err := b.Initialize(context.Background()); err != nil {
		return err
	}
	return b.Run(context.Background())
}
