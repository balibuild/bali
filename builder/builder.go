package builder

// Builder build syso file
type Builder struct {
}

// NewBuilder create builder
func NewBuilder(arch string) (*Builder, error) {
	b := &Builder{}

	return b, nil
}

// AddIcon add icon to resources
func (b *Builder) AddIcon(src string) error {

	return nil
}

// AddManifest todo
func (b *Builder) AddManifest(src string) error {

	return nil
}
