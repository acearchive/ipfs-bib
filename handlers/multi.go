package handlers

import "context"

type MultiHandler []DownloadHandler

func (m MultiHandler) Handle(ctx context.Context, response *HttpResponse) (*SourceContent, error) {
	for _, handler := range m {
		content, err := handler.Handle(ctx, response)

		switch {
		case err != nil:
			return nil, err
		case content != nil:
			return content, nil
		}
	}

	return nil, nil
}
