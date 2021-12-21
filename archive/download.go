package archive

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/handler"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/frawleyskid/ipfs-bib/resolver"
	"io"
	"net/http"
)

type Client struct {
	httpClient *network.HttpClient
}

func NewClient(httpClient *network.HttpClient) *Client {
	return &Client{httpClient}
}

func (c Client) Download(ctx context.Context, locator *config.SourceLocator, downloadHandler handler.DownloadHandler, sourceResolver resolver.SourceResolver) (*handler.SourceContent, error) {
	redirectedUrl, err := c.httpClient.ResolveRedirect(ctx, &locator.Url)
	if err != nil {
		return nil, err
	}

	redirectedLocator := &config.SourceLocator{
		Doi: locator.Doi,
		Url: *redirectedUrl,
	}

	resolvedUrl, err := sourceResolver.Resolve(ctx, redirectedLocator)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Request(ctx, http.MethodGet, resolvedUrl)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := response.Body.Close(); err != nil {
		return nil, err
	}

	downloadResponse := &handler.DownloadResponse{
		Url:    *resolvedUrl,
		Header: response.Header,
		Body:   responseBody,
	}

	content, err := downloadHandler.Handle(ctx, downloadResponse)
	if err != nil {
		return nil, err
	}

	return content, nil
}
