package store

import (
	ipfs "github.com/ipfs/go-ipfs-http-client"
	"github.com/multiformats/go-multiaddr"
)

func IpfsClient(apiAddr string) (*ipfs.HttpApi, error) {
	multiAddr, err := multiaddr.NewMultiaddr(apiAddr)
	if err != nil {
		return nil, err
	}

	ipfsApi, err := ipfs.NewApi(multiAddr)
	if err != nil {
		return nil, err
	}

	return ipfsApi, nil
}