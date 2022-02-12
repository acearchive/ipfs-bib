package resolver

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/network"
	"net/url"
)

var ErrNotResolved = errors.New("content not resolved")

type ContentOrigin string

const ContentOriginUrl ContentOrigin = "url"

type ResolvedLocator struct {
	OriginalUrl   url.URL
	ResolvedUrl   url.URL
	Origin        ContentOrigin
	MediaTypeHint *string
}

type SourceResolver interface {
	Resolve(ctx context.Context, locator config.SourceLocator) (ResolvedLocator, error)
}

type DirectResolver struct{}

func (DirectResolver) Resolve(_ context.Context, locator config.SourceLocator) (ResolvedLocator, error) {
	return ResolvedLocator{
		ResolvedUrl:   locator.Url,
		OriginalUrl:   locator.Url,
		Origin:        ContentOriginUrl,
		MediaTypeHint: nil,
	}, nil
}

type NoOpResolver struct{}

func (NoOpResolver) Resolve(_ context.Context, _ config.SourceLocator) (ResolvedLocator, error) {
	return ResolvedLocator{}, ErrNotResolved
}

func FromConfig(cfg config.Config) (SourceResolver, error) {
	userResolver, err := NewUserResolver(network.NewClient(cfg.File.Archive.UserAgent), cfg.File.Resolvers)
	if err != nil {
		return nil, err
	}

	return MultiResolver{
		NewUnpaywallResolver(network.NewClient(cfg.File.Archive.UserAgent), cfg),
		userResolver,
		DirectResolver{},
	}, nil
}
