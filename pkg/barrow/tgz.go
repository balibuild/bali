package barrow

import "context"

func (b *BarrowCtx) stgz(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (b *BarrowCtx) tgz(ctx context.Context, p *Package, crates []*Crate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
