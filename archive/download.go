package archive

import (
	"context"
	"github.com/go-shiori/obelisk"
	"io"
	"net/http"
	"net/url"
	"time"
)

const DefaultTimeout time.Duration = 1000 * 1000 * 1000 * 15

var DefaultUserAgent = obelisk.DefaultUserAgent

type DownloadClient struct {
	httpClient http.Client
}

func NewDownloadClient() *DownloadClient {
	return &DownloadClient{
		httpClient: http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func (c *DownloadClient) request(method string, url *url.URL) (*HttpResponse, error) {
	request, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set(UserAgentHeader, DefaultUserAgent)

	response, err := c.httpClient.Do(request)
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

	return &HttpResponse{
		Url:    *url,
		Body:   content,
		Header: response.Header,
	}, nil
}

func (c *DownloadClient) Download(ctx context.Context, url *url.URL) (content *SourceContent, filename string, err error) {
	response, err := c.request("GET", url)
	if err != nil {
		return nil, "", err
	}

	content, err = DefaultHandler.Handle(ctx, response)
	if err != nil {
		return nil, "", err
	}

	filename, err = InferFileName(response.ContentDisposition(), content.MediaType)
	if err != nil {
		return nil, "", err
	}

	return content, filename, err
}
