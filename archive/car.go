package archive

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2/blockstore"
)

func WriteCar(ctx context.Context, path string, store SourceStore) error {
	car, err := blockstore.OpenReadWrite(path, []cid.Cid{store.node.Cid()})
	if err != nil {
		return err
	}

	if err := car.Put(ctx, store.node); err != nil {
		return err
	}

	return car.Finalize()
}
