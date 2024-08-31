package main

import (
	"github.com/balibuild/bali/v3/pkg/barrow"
)

type CleanCommand struct {
	Force       bool   `name:"force" help:"Clean up all the builds"`
	Destination string `name:"destination" short:"D" help:"Specify the package save destination" default:"out"`
}

func (c *CleanCommand) Run(g *Globals) error {
	b := barrow.BarrowCtx{
		CWD:         g.M,
		Out:         g.B,
		Destination: c.Destination,
		Verbose:     g.Verbose,
	}
	return b.Cleanup(c.Force)
}
