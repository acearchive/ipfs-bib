package archive

import (
	"github.com/ipfs/go-merkledag"
	"io"
)

type SourceStore struct {
	node *merkledag.ProtoNode
}

func (d SourceStore) AddSource(source *BibSource) (id *BibSourceId, err error) {
	defer func() {
		err = source.Content.Close()
	}()

	sourceContentData, err := io.ReadAll(source.Content)
	if err != nil {
		return nil, err
	}

	sourceDirectoryNode := merkledag.NodeWithData(nil)
	sourceContentNode := merkledag.NewRawNode(sourceContentData)

	if err := sourceDirectoryNode.AddNodeLink(source.FileName, sourceContentNode); err != nil {
		return nil, err
	}

	if err := d.node.AddNodeLink(source.DirectoryName, sourceDirectoryNode); err != nil {
		return nil, err
	}

	return &BibSourceId{
		ContentCid:   sourceContentNode.Cid(),
		DirectoryCid: sourceDirectoryNode.Cid(),
	}, err
}
