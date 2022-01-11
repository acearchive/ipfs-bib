package store

import (
	"bytes"
	"context"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/ipfs/go-cid"
	chunk "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer"
	unixfs "github.com/ipfs/go-unixfs/io"
)

type dagSourceStore struct {
	service   ipld.DAGService
	directory unixfs.Directory
}

func newDagSourceStore(ctx context.Context, service ipld.DAGService) (*dagSourceStore, error) {
	directory := unixfs.NewDirectory(service)
	directory.SetCidBuilder(DefaultCidPrefix)

	dirNode, err := directory.GetNode()
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	if err := service.Add(ctx, dirNode); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	return &dagSourceStore{
		service:   service,
		directory: directory,
	}, nil
}

func (s *dagSourceStore) AddSource(ctx context.Context, source config.BibSource) (config.BibEntryLocation, error) {
	contentNode, err := importer.BuildDagFromReader(s.service, chunk.DefaultSplitter(bytes.NewReader(source.Content)))
	if err != nil {
		return config.BibEntryLocation{}, fmt.Errorf("%w: %v", ErrIpfs, err)
	}

	sourceDirectory := unixfs.NewDirectory(s.service)
	sourceDirectory.SetCidBuilder(DefaultCidPrefix)

	if err := sourceDirectory.AddChild(ctx, source.FileName, contentNode); err != nil {
		return config.BibEntryLocation{}, err
	}

	directoryNode, err := sourceDirectory.GetNode()
	if err != nil {
		return config.BibEntryLocation{}, fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	if err := s.service.Add(ctx, directoryNode); err != nil {
		return config.BibEntryLocation{}, fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	if err := s.directory.AddChild(ctx, source.DirectoryName, directoryNode); err != nil {
		return config.BibEntryLocation{}, fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	return config.BibEntryLocation{
		FileCid:       contentNode.Cid(),
		FileName:      source.FileName,
		DirectoryCid:  directoryNode.Cid(),
		DirectoryName: source.DirectoryName,
	}, nil
}

func (s *dagSourceStore) Finalize(ctx context.Context) (cid.Cid, error) {
	node, err := s.directory.GetNode()
	if err != nil {
		return cid.Undef, fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	if err := s.service.Add(ctx, node); err != nil {
		return cid.Undef, fmt.Errorf("%w, %v", ErrIpfs, err)
	}

	return node.Cid(), nil
}
