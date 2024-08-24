package main

type BuildCommand struct {
	Target string `name:"target" short:"T" help:"Target OS for which the code is compiled" default:"${target}"`       // windows/darwin
	Arch   string `name:"arch" short:"A" help:"Target architecture for which the code is compiled" default:"${arch}"` // amd64/arm64 ...
}

func (c *BuildCommand) Run(g *Globals) error {

	return nil
}
