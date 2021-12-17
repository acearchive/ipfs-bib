package archive

import (
	"bytes"
	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
)

type Node struct {
	node *shell.Shell
}

func NewNode(url string) Node {
	return Node{node: shell.NewShell(url)}
}

func (n Node) WriteSources(store SourceStore, pin bool) (id cid.Cid, err error) {
	encodedNode, err := store.node.EncodeProtobuf(false)
	if err != nil {
		return cid.Undef, err
	}

	rawCid, err := n.node.Add(bytes.NewReader(encodedNode), shell.Pin(pin))

	id, err = cid.Decode(rawCid)
	if err != nil {
		return cid.Undef, err
	}

	return id, err
}
