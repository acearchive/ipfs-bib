package store

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/frawleyskid/ipfs-bib/config"
)

type NullSourceStore struct {
	store   *dagSourceStore
	service *LocalService
}

func NewNullSourceStore(ctx context.Context) (*NullSourceStore, error) {
	service, err := NewLocalService()
	if err != nil {
		return nil, err
	}

	store, err := newDagSourceStore(ctx, service)
	if err != nil {
		return nil, err
	}

	return &NullSourceStore{
		store:   store,
		service: service,
	}, nil
}

func (s *NullSourceStore) AddSource(ctx context.Context, source config.BibSource) (config.BibEntryLocation, error) {
	return s.store.AddSource(ctx, source)
}

func (s *NullSourceStore) Finalize(ctx context.Context) (cid.Cid, error) {
	rootCid, err := s.store.Finalize(ctx)
	if err != nil {
		return cid.Undef, err
	}

	if err := s.service.Close(); err != nil {
		return cid.Undef, err
	}

	return rootCid, nil
}
