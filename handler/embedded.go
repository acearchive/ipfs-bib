package handler

import (
	"bytes"
	"context"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"net/http"
)

type EmbeddedHandler struct {
	tagFinder *TagFinder
}

func NewEmbeddedHandler(mediaTypes []string) DownloadHandler {
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
	}
}

func (e *EmbeddedHandler) Handle(_ context.Context, response *DownloadResponse) (*SourceContent, error) {
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
		var contentUrl string

		switch embeddedNode.DataAtom {
		case atom.Object:
			if value := FindAttr(embeddedNode, "data"); value != nil {
				contentUrl = *value
			} else {
				return nil, nil
			}
		case atom.Embed:
			if value := FindAttr(embeddedNode, "src"); value != nil {
				contentUrl = *value
			} else {
				return nil, nil
			}
			contentUrl = *FindAttr(embeddedNode, "src")
		default:
			panic("unexpected html node")
		}

		response, err := http.Get(contentUrl)
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
