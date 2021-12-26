package resolver

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
)

type MultiResolver []SourceResolver

func (m MultiResolver) Resolve(ctx context.Context, locator config.SourceLocator) (ResolvedLocator, error) {
	for _, resolver := range m {
		resolvedLocator, err := resolver.Resolve(ctx, locator)

		switch {
		case errors.Is(err, ErrNotResolved):
			continue
		case err != nil:
			logging.Verbose.Println(err)
			continue
		}

		return resolvedLocator, nil
	}

	return ResolvedLocator{}, ErrNotResolved
}
