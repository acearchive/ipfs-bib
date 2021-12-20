package handler

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"mime"
	"net/http"
	"net/url"
)

const (
	ContentTypeHeader        = "Content-Type"
	ContentDispositionHeader = "Content-Disposition"
	DefaultMediaType         = "application/octet-stream"
)

type SourceContent struct {
	Content   []byte
	MediaType string
}

type DownloadResponse struct {
	Url    url.URL
	Body   []byte
	Header http.Header
}

func (r *DownloadResponse) MediaType() string {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(ContentTypeHeader))
	if err == nil {
		return mediaType
	} else {
		return DefaultMediaType
	}
}

func (r *DownloadResponse) ContentDisposition() string {
	return r.Header.Get(ContentDispositionHeader)
}

type DownloadHandler interface {
	Handle(ctx context.Context, response *DownloadResponse) (*SourceContent, error)
}

type PassthroughHandler struct{}

func (s *PassthroughHandler) Handle(_ context.Context, response *DownloadResponse) (*SourceContent, error) {
	return &SourceContent{
		Content:   response.Body,
		MediaType: response.MediaType(),
	}, nil
}

type NoOpHandler struct{}

func (n *NoOpHandler) Handle(_ context.Context, _ *DownloadResponse) (*SourceContent, error) {
	return nil, nil
}

func FromConfig(cfg *config.Config) DownloadHandler {
	return MultiHandler{
		NewEmbeddedHandler(cfg.Archive.EmbeddedTypes),
		NewMonolithHandler(&cfg.Snapshot),
		&PassthroughHandler{},
	}
}
