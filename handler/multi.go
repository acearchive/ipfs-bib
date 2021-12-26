package handler

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/logging"
)

type MultiHandler []DownloadHandler

func (m MultiHandler) Handle(ctx context.Context, response DownloadResponse) (SourceContent, error) {
	for _, handler := range m {
		content, err := handler.Handle(ctx, response)

		switch {
		case errors.Is(err, ErrNotHandled):
			continue
		case err != nil:
			logging.Verbose.Println(err)
			continue
		}

		return content, nil
	}

	return SourceContent{}, ErrNotHandled
}
