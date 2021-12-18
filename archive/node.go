package archive

import (
	"context"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	unixfs "github.com/ipfs/go-unixfs/io"
	"io"
)

type NodeStore struct {
	service   ipld.DAGService
	directory unixfs.Directory
}

func NewNodeStore(ctx context.Context, service ipld.DAGService) (*NodeStore, error) {
	directory := unixfs.NewDirectory(service)

	dirNode, err := directory.GetNode()
	if err != nil {
		return nil, err
	}

	if err := service.Add(ctx, dirNode); err != nil {
		return nil, err
	}

	return &NodeStore{service, directory}, nil
}

func (s *NodeStore) Write(ctx context.Context) (cid.Cid, error) {
	node, err := s.directory.GetNode()
	if err != nil {
		return cid.Undef, err
	}

	if err := s.service.Add(ctx, node); err != nil {
		return cid.Undef, err
	}

	return node.Cid(), nil
}

func (s *NodeStore) AddSource(ctx context.Context, source *BibSource) (id *BibSourceId, err error) {
	defer func() {
		err = source.Content.Close()
	}()

	sourceData, err := io.ReadAll(source.Content)
	if err != nil {
		return nil, err
	}

	contentNode := dag.NewRawNode(sourceData)
	if err := s.service.Add(ctx, contentNode); err != nil {
		return nil, err
	}

	sourceDirectory := unixfs.NewDirectory(s.service)

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

	return &BibSourceId{
		ContentCid:   contentNode.Cid(),
		DirectoryCid: directoryNode.Cid(),
	}, err
}
