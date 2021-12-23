package resolver

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
)

type MultiResolver []SourceResolver

func (m MultiResolver) Resolve(ctx context.Context, locator *config.SourceLocator) (*ResolvedLocator, error) {
	for _, resolver := range m {
		resolvedLocator, err := resolver.Resolve(ctx, locator)

		switch {
		case err != nil:
			return nil, err
		case resolvedLocator != nil:
			return resolvedLocator, nil
		}
	}

	return nil, nil
}
