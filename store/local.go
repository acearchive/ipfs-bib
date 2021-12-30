package store

import (
	"github.com/ipfs/go-blockservice"
	syncds "github.com/ipfs/go-datastore/sync"
	flatfs "github.com/ipfs/go-ds-flatfs"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	"io/ioutil"
	"os"
)

type LocalService struct {
	ipld.DAGService
	ds   *flatfs.Datastore
	path string
}

func NewLocalService() (*LocalService, error) {
	storePath, err := ioutil.TempDir("", "ipfs-bib-ds-*")
	if err != nil {
		return nil, err
	}

	if err := flatfs.Create(storePath, flatfs.IPFS_DEF_SHARD); err != nil {
		return nil, err
	}

	datastore, err := flatfs.Open(storePath, false)
	if err != nil {
		return nil, err
	}

	store := blockstore.NewBlockstore(syncds.MutexWrap(datastore))
	blockService := blockservice.New(store, offline.Exchange(store))

	service := dag.NewDAGService(blockService)

	return &LocalService{
		DAGService: service,
		ds:         datastore,
		path:       storePath,
	}, nil
}

func (s *LocalService) Close() error {
	if err := s.ds.Close(); err != nil {
		return err
	}

	return os.RemoveAll(s.path)
}
