package archive

import (
	"bytes"
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/handlers"
	"io"
	"net/http"
	"net/url"
	"text/template"
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

func (c *DownloadClient) Download(ctx context.Context, sourceUrl *url.URL, handler handlers.DownloadHandler) (content *handlers.SourceContent, filename string, err error) {
	response, err := c.request(http.MethodGet, sourceUrl)
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

func (c *DownloadClient) ResolveProxy(locator *config.SourceLocator, cfg []config.Proxy) ([]url.URL, error) {
	// If the URL is an https://doi.org/ URL, we need to find the URL it
	// redirects to before passing it to the template.
	response, err := c.request(http.MethodHead, &locator.Url)
	if err != nil {
		return nil, err
	}

	resolvedLocator := &config.ResolvedSourceLocator{
		Doi: locator.Doi,
		Url: response.Url,
	}

	var proxiedUrls []url.URL

	for _, proxyCfg := range cfg {
		templateInput, err := config.NewProxySchemeInput(resolvedLocator, &proxyCfg)
		if err != nil {
			return nil, err
		} else if templateInput == nil {
			// This source was excluded by the hostname include/exclude rules.
			continue
		}

		for _, scheme := range proxyCfg.Schemes {
			schemeTemplate, err := template.New("proxy-scheme").Parse(scheme)
			if err != nil {
				return nil, err
			}

			var rawProxyUrlBytes []byte

			if err := schemeTemplate.Execute(bytes.NewBuffer(rawProxyUrlBytes), templateInput); err != nil {
				return nil, err
			}

			if len(rawProxyUrlBytes) == 0 {
				// We skip templates that resolve to an empty string.
				continue
			}

			rawProxyUrl := string(rawProxyUrlBytes)

			proxyUrl, err := url.Parse(rawProxyUrl)
			if err != nil {
				return nil, err
			}

			proxiedUrls = append(proxiedUrls, *proxyUrl)
		}
	}

	return proxiedUrls, nil
}
