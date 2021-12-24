package handler

import (
	"bytes"
	"context"
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
		tagFinder: NewTagFinder(func(node *html.Node) bool {
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

func (e *EmbeddedHandler) Handle(ctx context.Context, response *DownloadResponse) (*SourceContent, error) {
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
		var rawContentUrl string

		switch embeddedNode.DataAtom {
		case atom.Object:
			if value := FindAttr(embeddedNode, "data"); value != nil {
				rawContentUrl = *value
			} else {
				return nil, nil
			}
		case atom.Embed:
			if value := FindAttr(embeddedNode, "src"); value != nil {
				rawContentUrl = *value
			} else {
				return nil, nil
			}
		default:
			panic("unexpected html node")
		}

		contentUrl, err := url.Parse(rawContentUrl)
		if err != nil {
			return nil, err
		}

		if contentUrl.Scheme == "" {
			contentUrl.Scheme = DefaultUrlScheme
		}

		response, err := e.httpClient.Request(ctx, http.MethodGet, contentUrl)
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

		mediaType := FindAttr(embeddedNode, "type")
		if mediaType == nil {
			panic("node unexpectedly missing its content type")
		}

		return &SourceContent{
			Content:   content,
			MediaType: *mediaType,
		}, nil
	}

	return nil, nil
}
