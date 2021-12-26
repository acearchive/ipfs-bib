package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"net/http"
	"net/url"
)

const DefaultUrlScheme = "https"

type EmbeddedHandler struct {
	tagFinder  *TagFinder
	httpClient *network.HttpClient
}

func NewEmbeddedHandler(userAgent string, mediaTypes []string) DownloadHandler {
	if len(mediaTypes) == 0 {
		return &NoOpHandler{}
	}

	mediaTypeSet := make(map[string]struct{})
	for _, mediaType := range mediaTypes {
		mediaTypeSet[mediaType] = struct{}{}
	}

	return &EmbeddedHandler{
		tagFinder: NewTagFinder(func(node html.Node) bool {
			if node.Type == html.ElementNode && (node.DataAtom == atom.Object || node.DataAtom == atom.Embed) {
				if value := FindAttr(node, "type"); value != nil {
					_, exists := mediaTypeSet[*value]
					return exists
				}
			}

			return false
		}),
		httpClient: network.NewClient(userAgent),
	}
}

func (e *EmbeddedHandler) Handle(ctx context.Context, response DownloadResponse) (SourceContent, error) {
	if response.MediaType() != "text/html" {
		return SourceContent{}, ErrNotHandled
	}

	rootNode, err := html.Parse(bytes.NewReader(response.Body))
	if err != nil {
		return SourceContent{}, fmt.Errorf("%w: %v", network.ErrUnmarshalResponse, err)
	}

	documentNode := FindChild(*rootNode, func(node html.Node) bool {
		return node.Type == html.ElementNode && node.DataAtom == atom.Html
	})

	if documentNode == nil {
		return SourceContent{}, ErrNotHandled
	}

	var embeddedNode html.Node

	if node := e.tagFinder.Find(documentNode); node != nil {
		embeddedNode = *node
	} else {
		return SourceContent{}, ErrNotHandled
	}

	var rawContentUrl string

	switch embeddedNode.DataAtom {
	case atom.Object:
		if value := FindAttr(embeddedNode, "data"); value != nil {
			rawContentUrl = *value
		} else {
			return SourceContent{}, ErrNotHandled
		}
	case atom.Embed:
		if value := FindAttr(embeddedNode, "src"); value != nil {
			rawContentUrl = *value
		} else {
			return SourceContent{}, ErrNotHandled
		}
	default:
		logging.Error.Fatal("unexpected HTML node type")
	}

	contentUrl, err := url.Parse(rawContentUrl)
	if err != nil {
		return SourceContent{}, fmt.Errorf("%w: %v", network.ErrUnmarshalResponse, err)
	}

	if contentUrl.Scheme == "" {
		contentUrl.Scheme = DefaultUrlScheme
	}

	embeddedResponse, err := e.httpClient.Request(ctx, http.MethodGet, *contentUrl)
	if err != nil {
		return SourceContent{}, err
	}

	content, err := io.ReadAll(embeddedResponse.Body)
	if err != nil {
		return SourceContent{}, fmt.Errorf("%w: %v", network.ErrHttp, err)
	}

	if err := embeddedResponse.Body.Close(); err != nil {
		return SourceContent{}, fmt.Errorf("%w: %v", network.ErrHttp, err)
	}

	var mediaType string

	if maybeMediaType := FindAttr(embeddedNode, "type"); maybeMediaType != nil {
		mediaType = *maybeMediaType
	} else {
		logging.Error.Fatal("node unexpectedly missing its content type")
	}

	return SourceContent{
		Content:   content,
		MediaType: mediaType,
		FileName:  config.InferFileName(contentUrl, mediaType, embeddedResponse.Header),
	}, nil
}
