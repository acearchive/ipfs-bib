package store

import (
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/mount"
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

	fsStore, err := flatfs.Open(storePath, false)
	if err != nil {
		return nil, err
	}

	mountStore := mount.New([]mount.Mount{{
		Prefix:    datastore.NewKey("/blocks"),
		Datastore: fsStore,
	}})

	blockStore := blockstore.NewBlockstore(syncds.MutexWrap(mountStore))
	blockService := blockservice.New(blockStore, offline.Exchange(blockStore))

	service := dag.NewDAGService(blockService)

	return &LocalService{
		DAGService: service,
		ds:         fsStore,
		path:       storePath,
	}, nil
}

func (s *LocalService) Close() error {
	if err := s.ds.Close(); err != nil {
		return err
	}

	return os.RemoveAll(s.path)
}
