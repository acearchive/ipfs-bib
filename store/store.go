package store

import (
	"context"
	"errors"
	"github.com/ipfs/go-cid"
	dag "github.com/ipfs/go-merkledag"
	"github.com/frawleyskid/ipfs-bib/config"
)

var DefaultCidPrefix = dag.V1CidPrefix()

var ErrIpfs = errors.New("ipfs error")

type SourceStore interface {
	AddSource(ctx context.Context, source config.BibSource) (config.BibEntryLocation, error)
	Finalize(ctx context.Context) (cid.Cid, error)
}

func SourceStoreFromConfig(ctx context.Context, cfg config.Config) (SourceStore, error) {
	switch {
	case cfg.Flags.DryRun:
		return NewNullSourceStore(ctx)
	case cfg.Flags.MaybeCarPath() == nil:
		options := NodeSourceStoreOptions{
			PinLocal:        cfg.Flags.PinLocal,
			PinRemoteName:   cfg.Flags.MaybePinRemoteName(),
			PinningServices: cfg.File.Pins,
			MfsPath:         cfg.Flags.MaybeMfsPath(),
		}

		return NewNodeSourceStore(ctx, cfg.File.Ipfs.Api, options)
	default:
		isCarV2, err := cfg.File.Ipfs.IsCarV2()
		if err != nil {
			return nil, err
		}

		return NewCarSourceStore(ctx, cfg.Flags.CarPath, isCarV2)
	}
}
