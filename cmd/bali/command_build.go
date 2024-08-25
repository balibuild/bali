package main

type BuildCommand struct {
	Target  string `name:"target" short:"T" help:"Target OS for which the code is compiled" default:"${target}"`       // windows/darwin
	Arch    string `name:"arch" short:"A" help:"Target architecture for which the code is compiled" default:"${arch}"` // amd64/arm64 ...
	Release string `name:"release" help:"Specifies the rpm package tag version"`                                       // --release $TASK_ID
	Pack    string `name:"pack" help:"Pack in the specified format after the build is complete"`
}

func (c *BuildCommand) Run(g *Globals) error {

	return nil
}
