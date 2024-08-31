package main

type UpdateCommand struct {
	ALL     bool     `name:"all" help:"Update all dependencies"`
	Modules []string `name:"modules" short:"m" help:"Update modules to the specified version"`
}

func (c *UpdateCommand) Run(g *Globals) error {

	return nil
}
