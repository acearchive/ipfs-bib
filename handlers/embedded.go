package handlers

import (
	"bytes"
	"context"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"net/http"
)

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
