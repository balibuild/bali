package pecoff

import "github.com/akavel/rsrc/coff"

// Builder build syso file
type Builder struct {
	c *coff.Coff
}

// NewBuilder create builder
func NewBuilder(arch string) (*Builder, error) {
	builder := &Builder{c: coff.NewRSRC()}
	if err := builder.c.Arch(arch); err != nil {
		return nil, err
	}
	return builder, nil
}
