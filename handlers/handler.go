package handlers

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

type HttpResponse struct {
	Url    url.URL
	Body   []byte
	Header http.Header
}

type SourceContent struct {
	Content   []byte
	MediaType string
}

func (r *HttpResponse) MediaType() string {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(ContentTypeHeader))
	if err == nil {
		return mediaType
	} else {
		return DefaultMediaType
	}
}

func (r *HttpResponse) ContentDisposition() string {
	return r.Header.Get(ContentDispositionHeader)
}

type DownloadHandler interface {
	Handle(ctx context.Context, response *HttpResponse) (*SourceContent, error)
}

type PassthroughHandler struct{}

func (s *PassthroughHandler) Handle(_ context.Context, response *HttpResponse) (*SourceContent, error) {
	return &SourceContent{
		Content:   response.Body,
		MediaType: response.MediaType(),
	}, nil
}

type NoOpHandler struct{}

func (n *NoOpHandler) Handle(_ context.Context, response *HttpResponse) (*SourceContent, error) {
	return nil, nil
}

func FromConfig(cfg *config.Handlers) DownloadHandler {
	return MultiHandler{
		NewEmbeddedHandler(&cfg.Embed),
		NewMonolithHandler(&cfg.Monolith),
		&PassthroughHandler{},
	}
}
