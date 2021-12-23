package resolver

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/network"
	"net/url"
)

type ContentOrigin string

const ContentOriginUrl ContentOrigin = "url"

type ResolvedLocator struct {
	Url    url.URL
	Origin ContentOrigin
}

type SourceResolver interface {
	Resolve(ctx context.Context, locator *config.SourceLocator) (*ResolvedLocator, error)
}

type DirectResolver struct{}

func (DirectResolver) Resolve(_ context.Context, locator *config.SourceLocator) (*ResolvedLocator, error) {
	return &ResolvedLocator{Url: locator.Url, Origin: ContentOriginUrl}, nil
}

type NoOpResolver struct{}

func (NoOpResolver) Resolve(_ context.Context, _ *config.SourceLocator) (*ResolvedLocator, error) {
	return nil, nil
}

func FromConfig(cfg *config.Config) (SourceResolver, error) {
	userResolver, err := NewUserResolver(network.NewClient(cfg.Archive.UserAgent), cfg.Resolvers)
	if err != nil {
		return nil, err
	}

	return &MultiResolver{
		NewUnpaywallResolver(network.NewClient(cfg.Archive.UserAgent), cfg),
		userResolver,
		DirectResolver{},
	}, nil
}
