package handler

import (
	"context"
	"errors"
	"github.com/frawleyskid/ipfs-bib/config"
	"mime"
	"net/http"
	"net/url"
)

const ContentTypeHeader = "Content-Type"

var ErrNotHandled = errors.New("handler could not handle content")

type SourceContent struct {
	Content   []byte
	MediaType string
	FileName  string
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
		return config.DefaultMediaType
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
		FileName:  config.InferFileName(&response.Url, response.MediaType(), response.Header),
	}, nil
}

type NoOpHandler struct{}

func (n *NoOpHandler) Handle(_ context.Context, _ *DownloadResponse) (*SourceContent, error) {
	return nil, ErrNotHandled
}

func FromConfig(cfg *config.Config) DownloadHandler {
	var excludeTypes []string

	if !cfg.Snapshot.Enabled {
		// If taking snapshots is disabled, don't attempt to download HTML documents.
		excludeTypes = []string{"text/html"}
	}

	return MultiHandler{
		NewEmbeddedHandler(cfg.Archive.UserAgent, cfg.Archive.EmbeddedTypes),
		NewMonolithHandler(cfg),
		NewDirectHandler(excludeTypes),
	}
}
