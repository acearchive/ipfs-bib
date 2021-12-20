package resolver

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/network"
	"net/url"
)

type SourceResolver interface {
	Resolve(ctx context.Context, locator *config.SourceLocator) (*url.URL, error)
}

type PassthroughResolver struct{}

func (PassthroughResolver) Resolve(_ context.Context, locator *config.SourceLocator) (*url.URL, error) {
	return &locator.Url, nil
}

type NoOpResolver struct{}

func (NoOpResolver) Resolve(_ context.Context, _ *config.SourceLocator) (*url.URL, error) {
	return nil, nil
}

func FromConfig(cfg *config.Config) (SourceResolver, error) {
	userResolver, err := NewUserResolver(&network.DefaultClient, cfg.Resolvers)
	if err != nil {
		return nil, err
	}

	return &MultiResolver{
		userResolver,
		PassthroughResolver{},
	}, nil
}
