package store

import (
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	ipfs "github.com/ipfs/go-ipfs-http-client"
	pinning "github.com/ipfs/go-pinning-service-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	ipfspath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/frawleyskid/ipfs-bib/config"
)

func PinLocal(ctx context.Context, api *ipfs.HttpApi, id cid.Cid, recursive bool) error {
	ipfsPath := ipfspath.IpfsPath(id)
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

func PinRemote(ctx context.Context, id cid.Cid, name string, services []config.Pin) error {
	for _, service := range services {
		client := pinning.NewClient(service.Endpoint, service.Token)

		if _, err := client.Add(ctx, id, pinning.PinOpts.WithName(name)); err != nil {
			return err
		}
	}

	return nil
}

func AddToMfs(ctx context.Context, apiUrl string, cid cid.Cid, path string) error {
	api := shell.NewShell(apiUrl)
	return api.FilesCp(ctx, ipfspath.IpfsPath(cid).String(), path)
}
