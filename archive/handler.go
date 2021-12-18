package archive

import (
	"bytes"
	"context"
	"github.com/go-shiori/obelisk"
	"mime"
	"net/http"
	"net/url"
)

var DefaultHandler = MultiHandler{
	&EmbeddedPdfHandler{},
	NewSnapshotHandler(),
	&PassthroughHandler{},
}

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

type SnapshotHandler struct {
	archiver *obelisk.Archiver
}

func NewSnapshotHandler() *SnapshotHandler {
	archiver := obelisk.Archiver{
		UserAgent: obelisk.DefaultUserAgent,
	}

	archiver.Validate()

	return &SnapshotHandler{&archiver}
}

func (s *SnapshotHandler) Handle(ctx context.Context, response *HttpResponse) (*SourceContent, error) {
	if response.MediaType() != "text/html" {
		return nil, nil
	}

	content, contentType, err := s.archiver.Archive(ctx, obelisk.Request{
		URL:   response.Url.String(),
		Input: bytes.NewReader(response.Body),
	})
	if err != nil {
		return nil, err
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	return &SourceContent{
		Content:   content,
		MediaType: mediaType,
	}, nil
}

type EmbeddedPdfHandler struct{}

func (e *EmbeddedPdfHandler) Handle(_ context.Context, _ *HttpResponse) (*SourceContent, error) {
	// TODO: Implement
	return nil, nil
}

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
