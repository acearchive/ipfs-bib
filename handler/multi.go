package handler

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
)

type MultiHandler []DownloadHandler

func (m MultiHandler) Handle(ctx context.Context, response *DownloadResponse) (*SourceContent, error) {
	for _, handler := range m {
		content, err := handler.Handle(ctx, response)

		switch {
		case errors.Is(err, network.ErrHttp):
			logging.Verbose.Println(err)
		case err != nil:
			return nil, err
		case content != nil:
			return content, nil
		}
	}

	return nil, nil
}
