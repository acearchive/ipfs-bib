package store

import (
	"context"
	"errors"
	"fmt"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	gocar "github.com/ipld/go-car"
	gocarv2 "github.com/ipld/go-car/v2"
	selectorparse "github.com/ipld/go-ipld-prime/traversal/selector/parse"
	"github.com/frawleyskid/ipfs-bib/config"
	"io/ioutil"
	"os"
)

var (
	ErrCarAlreadyExists = errors.New("this file already exists")
)

type dagStore struct {
	service ipld.DAGService
}

func (ds dagStore) Get(ctx context.Context, c cid.Cid) (blocks.Block, error) {
	return ds.service.Get(ctx, c)
}

func writeCar(ctx context.Context, service ipld.DAGService, root cid.Cid, path string, carv2 bool) error {
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
		return fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	if err := carFile.Close(); err != nil {
		return err
	}

	if carv2 {
		if err := gocarv2.WrapV1File(carFile.Name(), path); err != nil {
			return fmt.Errorf("%w, %v", ErrIpfs, err)
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

type CarSourceStore struct {
	service *LocalService
	store   *dagSourceStore
	carPath string
	carv2   bool
}

func NewCarSourceStore(ctx context.Context, carPath string, carv2 bool) (*CarSourceStore, error) {
	service, err := NewLocalService()
	if err != nil {
		return nil, err
	}

	store, err := newDagSourceStore(ctx, service)
	if err != nil {
		return nil, err
	}

	return &CarSourceStore{
		service: service,
		store:   store,
		carPath: carPath,
		carv2:   carv2,
	}, nil
}

func (s *CarSourceStore) AddSource(ctx context.Context, source config.BibSource) (config.BibEntryLocation, error) {
	return s.store.AddSource(ctx, source)
}

func (s *CarSourceStore) Finalize(ctx context.Context) (cid.Cid, error) {
	rootCid, err := s.store.Finalize(ctx)
	if err != nil {
		return cid.Undef, err
	}

	if err := writeCar(ctx, s.service, rootCid, s.carPath, s.carv2); err != nil {
		return cid.Undef, err
	}

	if err := s.service.Close(); err != nil {
		return cid.Undef, err
	}

	return rootCid, nil
}
