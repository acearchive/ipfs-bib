package archive

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/handlers"
	"io"
	"net/http"
	"net/url"
	"time"
)

const DefaultTimeout time.Duration = 1000 * 1000 * 1000 * 15

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

func (c *DownloadClient) request(method string, requestUrl *url.URL) (*handlers.HttpResponse, error) {
	redirectedUrl := requestUrl

	c.httpClient.CheckRedirect = func(request *http.Request, _ []*http.Request) error {
		redirectedUrl = request.URL

		return nil
	}

	request, err := http.NewRequest(method, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

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

	return &handlers.HttpResponse{
		Url:    *redirectedUrl,
		Body:   content,
		Header: response.Header,
	}, nil
}

func (c *DownloadClient) Download(ctx context.Context, url *url.URL, handler handlers.DownloadHandler) (content *handlers.SourceContent, filename string, err error) {
	response, err := c.request("GET", url)
	if err != nil {
		return nil, "", err
	}

	content, err = handler.Handle(ctx, response)
	if err != nil {
		return nil, "", err
	}

	filename, err = InferFileName(response.ContentDisposition(), content.MediaType)
	if err != nil {
		return nil, "", err
	}

	return content, filename, err
}
