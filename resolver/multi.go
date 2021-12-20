package resolver

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"net/url"
)

type MultiResolver []SourceResolver

func (m MultiResolver) Resolve(ctx context.Context, locator *config.SourceLocator) (*url.URL, error) {
	for _, resolver := range m {
		resolvedUrl, err := resolver.Resolve(ctx, locator)

		switch {
		case err != nil:
			return nil, err
		case resolvedUrl != nil:
			return resolvedUrl, nil
		}
	}

	return nil, nil
}
