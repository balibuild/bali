package barrow

import "context"

func (b *BarrowCtx) zip(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
