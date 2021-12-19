package archive

import (
	"bytes"
	"context"
	"github.com/go-shiori/obelisk"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"mime"
	"net/http"
	"net/url"
)

var DefaultHandler = MultiHandler{
	NewEmbeddedPdfHandler(),
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

type EmbeddedPdfHandler struct {
	tagFinder *TagFinder
}

func NewEmbeddedPdfHandler() *EmbeddedPdfHandler {
	return &EmbeddedPdfHandler{
		tagFinder: NewTagFinder(func(node *html.Node) bool {
			if node.Type == html.ElementNode && (node.DataAtom == atom.Object || node.DataAtom == atom.Embed) {
				if value := FindAttr(node, "type"); value != nil {
					return *value == "application/pdf"
				}
			}

			return false
		}),
	}
}

func (e *EmbeddedPdfHandler) Handle(_ context.Context, response *HttpResponse) (*SourceContent, error) {
	if response.MediaType() != "text/html" {
		return nil, nil
	}

	rootNode, err := html.Parse(bytes.NewReader(response.Body))
	if err != nil {
		return nil, err
	}

	documentNode := FindChild(rootNode, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.DataAtom == atom.Html
	})

	if documentNode == nil {
		return nil, nil
	}

	if embeddedNode := e.tagFinder.Find(documentNode); embeddedNode != nil {
		var pdfUrl string

		switch embeddedNode.DataAtom {
		case atom.Object:
			if value := FindAttr(embeddedNode, "data"); value != nil {
				pdfUrl = *value
			} else {
				return nil, nil
			}
		case atom.Embed:
			if value := FindAttr(embeddedNode, "src"); value != nil {
				pdfUrl = *value
			} else {
				return nil, nil
			}
			pdfUrl = *FindAttr(embeddedNode, "src")
		default:
			panic("unexpected html node")
		}

		response, err := http.Get(pdfUrl)
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		if err := response.Body.Close(); err != nil {
			return nil, err
		}

		return &SourceContent{
			Content:   content,
			MediaType: "application/pdf",
		}, nil
	}

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
