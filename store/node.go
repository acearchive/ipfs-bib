package store

import (
	"context"
	"github.com/ipfs/go-cid"
	ipfs "github.com/ipfs/go-ipfs-http-client"
	"github.com/frawleyskid/ipfs-bib/config"
)

type NodeSourceStore struct {
	apiUrl          string
	ipfsApi         *ipfs.HttpApi
	store           *dagSourceStore
	pinLocal        bool
	pinRemoteName   *string
	pinningServices []config.Pin
	mfsPath         *string
}

type NodeSourceStoreOptions struct {
	PinLocal        bool
	PinRemoteName   *string
	PinningServices []config.Pin
	MfsPath         *string
}

func NewNodeSourceStore(ctx context.Context, apiUrl string, options NodeSourceStoreOptions) (*NodeSourceStore, error) {
	ipfsApi, err := IpfsClient(apiUrl)
	if err != nil {
		return nil, err
	}

	service := ipfsApi.Dag()

	store, err := newDagSourceStore(ctx, service)
	if err != nil {
		return nil, err
	}

	return &NodeSourceStore{
		apiUrl:          apiUrl,
		ipfsApi:         ipfsApi,
		store:           store,
		pinLocal:        options.PinLocal,
		pinRemoteName:   options.PinRemoteName,
		pinningServices: options.PinningServices,
		mfsPath:         options.MfsPath,
	}, nil
}

func (s *NodeSourceStore) AddSource(ctx context.Context, source config.BibSource) (config.BibEntryLocation, error) {
	return s.store.AddSource(ctx, source)
}

func (s *NodeSourceStore) Finalize(ctx context.Context) (cid.Cid, error) {
	rootCid, err := s.store.Finalize(ctx)
	if err != nil {
		return cid.Undef, err
	}

	if s.pinLocal {
		if err := PinLocal(ctx, s.ipfsApi, rootCid, true); err != nil {
			return cid.Undef, err
		}
	}

	if s.pinRemoteName != nil {
		if err := PinRemote(ctx, rootCid, *s.pinRemoteName, s.pinningServices); err != nil {
			return cid.Undef, err
		}
	}

	if s.mfsPath != nil {
		if err := AddToMfs(ctx, s.apiUrl, rootCid, *s.mfsPath); err != nil {
			return cid.Undef, err
		}
	}

	return rootCid, nil
}
