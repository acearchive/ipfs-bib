package archive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/handlers"
	"io"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

const DefaultTimeout time.Duration = 1000 * 1000 * 1000 * 15

type ErrHttpNotOk struct {
	Status     string
	StatusCode int
}

func (e ErrHttpNotOk) Error() string {
	return fmt.Sprintf("request returned with status code %s", e.Status)
}

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

	if ok := response.StatusCode >= 200 && response.StatusCode < 300; !ok {
		return nil, ErrHttpNotOk{StatusCode: response.StatusCode, Status: response.Status}
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

func (c *DownloadClient) downloadSource(ctx context.Context, sourceUrl *url.URL, handler handlers.DownloadHandler) (content *handlers.SourceContent, filename string, err error) {
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

func (c *DownloadClient) resolveProxy(locator *config.ResolvedSourceLocator, cfg []config.Proxy) ([]url.URL, error) {
	var proxiedUrls []url.URL

	for _, proxyCfg := range cfg {
		templateInput, err := config.NewProxySchemeInput(locator, &proxyCfg)
		if err != nil {
			return nil, err
		} else if templateInput == nil {
			// This source was excluded by the hostname include/exclude rules.
			continue
		}

		for _, scheme := range proxyCfg.Schemes {
			schemeTemplate, err := template.New("proxies.schemes").Funcs(sprig.TxtFuncMap()).Parse(scheme)
			if err != nil {
				return nil, err
			}

			var rawProxyUrlBytes bytes.Buffer

			if err := schemeTemplate.Execute(&rawProxyUrlBytes, templateInput); err != nil {
				return nil, err
			}

			if rawProxyUrlBytes.Len() == 0 {
				// We skip templates that resolve to an empty string.
				continue
			}

			rawProxyUrl := string(rawProxyUrlBytes.Bytes())

			proxyUrl, err := url.Parse(rawProxyUrl)
			if err != nil {
				return nil, err
			}

			proxiedUrls = append(proxiedUrls, *proxyUrl)
		}
	}

	return proxiedUrls, nil
}

func (c *DownloadClient) resolveLocator(locator *config.SourceLocator) (*config.ResolvedSourceLocator, error) {
	// If the URL is an https://doi.org/ URL, we need to find the URL it
	// redirects to before passing it to the template.
	response, err := c.request(http.MethodHead, &locator.Url)
	if err != nil {
		return nil, err
	}

	return &config.ResolvedSourceLocator{
		Doi: locator.Doi,
		Url: response.Url,
	}, nil
}

func (c *DownloadClient) DownloadWithProxy(ctx context.Context, locator *config.SourceLocator, cfg []config.Proxy, handler handlers.DownloadHandler) (content *handlers.SourceContent, filename string, err error) {
	resolvedLocator, err := c.resolveLocator(locator)
	if err != nil {
		return nil, "", err
	}

	proxiedUrls, err := c.resolveProxy(resolvedLocator, cfg)
	if err != nil {
		return nil, "", err
	}

	for _, proxyUrl := range proxiedUrls {
		content, filename, err := c.downloadSource(ctx, &proxyUrl, handler)
		if err != nil {
			httpErr := ErrHttpNotOk{}
			if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
				continue
			}

			return nil, "", err
		}

		return content, filename, nil
	}

	// There were no applicable proxy rules or all the proxy URLs 404'd.
	return c.downloadSource(ctx, &resolvedLocator.Url, handler)
}
