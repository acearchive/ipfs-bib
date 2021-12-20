package network

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const DefaultTimeout time.Duration = 1000 * 1000 * 1000 * 15

var DefaultClient = HttpClient{http.Client{
	Timeout: DefaultTimeout,
}}

type HttpClient struct {
	http.Client
}

func (c *HttpClient) Request(ctx context.Context, method string, requestUrl *url.URL) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.Do(request)
	if err != nil {
		return nil, err
	}

	if ok := response.StatusCode >= 200 && response.StatusCode < 300; !ok {
		return nil, ErrHttpNotOk{StatusCode: response.StatusCode, Status: response.Status}
	}

	return response, nil
}

func (c *HttpClient) ResolveRedirect(ctx context.Context, sourceUrl *url.URL) (*url.URL, error) {
	redirectedUrl := sourceUrl

	c.CheckRedirect = func(request *http.Request, _ []*http.Request) error {
		redirectedUrl = request.URL

		return nil
	}

	_, err := c.Request(ctx, http.MethodGet, sourceUrl)
	if err != nil {
		return nil, err
	}

	c.CheckRedirect = nil

	return redirectedUrl, nil
}

func (c *HttpClient) CheckExists(ctx context.Context, sourceUrl *url.URL) (bool, error) {
	_, err := c.Request(ctx, http.MethodGet, sourceUrl)
	if err != nil {
		httpErr := ErrHttpNotOk{}
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

type ErrHttpNotOk struct {
	Status     string
	StatusCode int
}

func (e ErrHttpNotOk) Error() string {
	return fmt.Sprintf("request returned with status code %s", e.Status)
}
