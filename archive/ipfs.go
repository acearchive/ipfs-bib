package archive

import (
	"context"
	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-merkledag"
	"github.com/ipld/go-car/v2/blockstore"
	"io"
)

func AddToNode(node shell.Shell, pin bool, content io.ReadCloser) (contentId cid.Cid, err error) {
	defer func() {
		err = content.Close()
	}()

	rawCid, err := node.Add(content, shell.Pin(pin))
	if err != nil {
		return cid.Undef, err
	}

	contentId, err = cid.Decode(rawCid)
	if err != nil {
		return cid.Undef, err
	}

	return contentId, err
}

func NewCar(path string) (*blockstore.ReadWrite, error) {
	return blockstore.OpenReadWrite(path, nil)
}

func AddToCar(ctx context.Context, car *blockstore.ReadWrite, content io.ReadCloser) (contentId cid.Cid, err error) {
	defer func() {
		err = content.Close()
	}()

	contentBuf, err := io.ReadAll(content)
	if err != nil {
		return cid.Undef, err
	}

	node := merkledag.NewRawNode(contentBuf)

	if err := car.Put(ctx, node.Block); err != nil {
		return cid.Undef, err
	}

	return node.Cid(), err
}
