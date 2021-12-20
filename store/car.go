package store

import (
	"context"
	"errors"
	"fmt"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	gocar "github.com/ipld/go-car"
	gocarv2 "github.com/ipld/go-car/v2"
	selectorparse "github.com/ipld/go-ipld-prime/traversal/selector/parse"
	"io/ioutil"
	"os"
)

var ErrCarAlreadyExists = errors.New("this file already exists")

type dagStore struct {
	service ipld.DAGService
}

func (ds dagStore) Get(ctx context.Context, c cid.Cid) (blocks.Block, error) {
	return ds.service.Get(ctx, c)
}

func CarService() (ipld.DAGService, error) {
	store := blockstore.NewBlockstore(syncds.MutexWrap(datastore.NewMapDatastore()))
	blockService := blockservice.New(store, offline.Exchange(store))
	return dag.NewDAGService(blockService), nil
}

func WriteCar(ctx context.Context, service ipld.DAGService, root cid.Cid, path string, carv2 bool) error {
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrCarAlreadyExists, path)
	}

	sourceDag := gocar.Dag{Root: root, Selector: selectorparse.CommonSelector_ExploreAllRecursively}
	store := dagStore{service}
	car := gocar.NewSelectiveCar(ctx, store, []gocar.Dag{sourceDag}, gocar.TraverseLinksOnlyOnce())

	carFile, err := ioutil.TempFile("", "ipfs-bib-*.car")
	if err != nil {
		return err
	}

	if err := car.Write(carFile); err != nil {
		return err
	}

	if err := carFile.Close(); err != nil {
		return err
	}

	if carv2 {
		if err := gocarv2.WrapV1File(carFile.Name(), path); err != nil {
			return err
		}

		if err := os.Remove(carFile.Name()); err != nil {
			return err
		}
	} else {
		if err := os.Rename(carFile.Name(), path); err != nil {
			return err
		}
	}

	return nil
}
