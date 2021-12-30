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
