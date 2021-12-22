package store

import (
	"context"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	unixfs "github.com/ipfs/go-unixfs/io"
	"github.com/frawleyskid/ipfs-bib/config"
)

var DefaultCidPrefix = dag.V1CidPrefix()

type SourceStore struct {
	service   ipld.DAGService
	directory unixfs.Directory
}

func NewSourceStore(ctx context.Context, service ipld.DAGService) (*SourceStore, error) {
	directory := unixfs.NewDirectory(service)
	directory.SetCidBuilder(DefaultCidPrefix)

	dirNode, err := directory.GetNode()
	if err != nil {
		return nil, err
	}

	if err := service.Add(ctx, dirNode); err != nil {
		return nil, err
	}

	return &SourceStore{service, directory}, nil
}

func (s *SourceStore) Write(ctx context.Context) (cid.Cid, error) {
	node, err := s.directory.GetNode()
	if err != nil {
		return cid.Undef, err
	}

	if err := s.service.Add(ctx, node); err != nil {
		return cid.Undef, err
	}

	return node.Cid(), nil
}

func (s *SourceStore) AddSource(ctx context.Context, source *config.BibSource) (*config.BibEntryLocation, error) {
	contentNode := dag.NewRawNode(source.Content)
	if err := s.service.Add(ctx, contentNode); err != nil {
		return nil, err
	}

	sourceDirectory := unixfs.NewDirectory(s.service)
	sourceDirectory.SetCidBuilder(DefaultCidPrefix)

	if err := sourceDirectory.AddChild(ctx, source.FileName, contentNode); err != nil {
		return nil, err
	}

	directoryNode, err := sourceDirectory.GetNode()
	if err != nil {
		return nil, err
	}

	if err := s.service.Add(ctx, directoryNode); err != nil {
		return nil, err
	}

	if err := s.directory.AddChild(ctx, source.DirectoryName, directoryNode); err != nil {
		return nil, err
	}

	return &config.BibEntryLocation{
		FileCid:       contentNode.Cid(),
		FileName:      source.FileName,
		DirectoryCid:  directoryNode.Cid(),
		DirectoryName: source.DirectoryName,
	}, nil
}
