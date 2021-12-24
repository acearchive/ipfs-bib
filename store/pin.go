package store

import (
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	ipfs "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
)

func Pin(ctx context.Context, api *ipfs.HttpApi, id cid.Cid, recursive bool) error {
	ipfsPath := path.IpfsPath(id)
	option := func(settings *options.PinAddSettings) error {
		settings.Recursive = recursive

		return nil
	}

	if err := api.Pin().Add(ctx, ipfsPath, option); err != nil {
		return fmt.Errorf("%w, %v", ErrIpfs, err)
	} else {
		return nil
	}
}
