package handler

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/network"
	"mime"
	"net/http"
	"net/url"
)

var ErrNotHandled = errors.New("handler could not handle content")

type SourceContent struct {
	Content   []byte
	MediaType string
	FileName  string
}

type DownloadResponse struct {
	Url           url.URL
	Body          []byte
	Header        http.Header
	MediaTypeHint *string
}

func (r DownloadResponse) MediaType() string {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(network.ContentTypeHeader))
	switch {
	case err != nil && r.MediaTypeHint != nil:
		return *r.MediaTypeHint
	case err != nil:
		return network.DefaultMediaType
	case mediaType == network.DefaultMediaType && r.MediaTypeHint != nil:
		return *r.MediaTypeHint
	}

	return mediaType
}

type DownloadHandler interface {
	Handle(ctx context.Context, response DownloadResponse) (SourceContent, error)
}

type DirectHandler struct {
	excludeTypes []string
}

func NewDirectHandler(excludeTypes []string) *DirectHandler {
	return &DirectHandler{excludeTypes}
}

func (s *DirectHandler) Handle(_ context.Context, response DownloadResponse) (SourceContent, error) {
	for _, mediaType := range s.excludeTypes {
		if response.MediaType() == mediaType {
			return SourceContent{}, ErrNotHandled
		}
	}

	return SourceContent{
		Content:   response.Body,
		MediaType: response.MediaType(),
		FileName:  config.InferFileName(&response.Url, response.MediaType(), response.Header),
	}, nil
}

type NoOpHandler struct{}

func (n *NoOpHandler) Handle(_ context.Context, _ DownloadResponse) (SourceContent, error) {
	return SourceContent{}, ErrNotHandled
}

func FromConfig(cfg config.Config) DownloadHandler {
	var excludeTypes []string

	if !cfg.File.Snapshot.Enabled {
		// If taking snapshots is disabled, don't attempt to download HTML documents.
		excludeTypes = []string{"text/html"}
	}

	return MultiHandler{
		NewEmbeddedHandler(cfg.File.Archive.UserAgent, cfg.File.Archive.EmbeddedTypes),
		NewMonolithHandler(cfg),
		NewDirectHandler(excludeTypes),
	}
}
