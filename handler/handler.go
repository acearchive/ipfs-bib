package handler

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"mime"
	"net/http"
	"net/url"
)

const (
	ContentTypeHeader = "Content-Type"
	DefaultMediaType  = "application/octet-stream"
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

type DownloadHandler interface {
	Handle(ctx context.Context, response *DownloadResponse) (*SourceContent, error)
}

type DirectHandler struct {
	excludeTypes []string
}

func NewDirectHandler(excludeTypes []string) *DirectHandler {
	return &DirectHandler{excludeTypes}
}

func (s *DirectHandler) Handle(_ context.Context, response *DownloadResponse) (*SourceContent, error) {
	for _, mediaType := range s.excludeTypes {
		if response.MediaType() == mediaType {
			return nil, nil
		}
	}

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
	var excludeTypes []string

	if !cfg.Snapshot.Enabled {
		// If taking snapshots is disabled, don't attempt to download HTML documents.
		excludeTypes = []string{"text/html"}
	}

	return MultiHandler{
		NewEmbeddedHandler(cfg.Archive.EmbeddedTypes),
		NewMonolithHandler(cfg),
		NewDirectHandler(excludeTypes),
	}
}
