package resolver

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
)

type MultiResolver []SourceResolver

func (m MultiResolver) Resolve(ctx context.Context, locator *config.SourceLocator) (*ResolvedLocator, error) {
	for _, resolver := range m {
		resolvedLocator, err := resolver.Resolve(ctx, locator)

		switch {
		case errors.Is(err, network.ErrHttp):
			logging.Verbose.Println(err)
		case err != nil:
			return nil, err
		case resolvedLocator != nil:
			return resolvedLocator, nil
		}
	}

	return nil, nil
}
