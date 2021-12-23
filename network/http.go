package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultTimeout  time.Duration = 1000 * 1000 * 1000 * 15
	UserAgentHeader string        = "User-Agent"
)

var defaultClient = http.Client{
	Timeout: DefaultTimeout,
}

func NewClient(userAgent string) *HttpClient {
	return &HttpClient{
		client:    &defaultClient,
		userAgent: userAgent,
	}
}

type HttpClient struct {
	client    *http.Client
	userAgent string
}

func (c *HttpClient) Request(ctx context.Context, method string, requestUrl *url.URL) (*http.Response, error) {
	return c.RequestWithHeaders(ctx, method, requestUrl, nil)
}

func (c *HttpClient) RequestWithHeaders(ctx context.Context, method string, requestUrl *url.URL, headers map[string]string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set(UserAgentHeader, c.userAgent)

	for headerName, headerValue := range headers {
		request.Header.Set(headerName, headerValue)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	if ok := response.StatusCode >= 200 && response.StatusCode < 300; !ok {
		return nil, ErrHttpNotOk{
			Method:     method,
			Url:        *requestUrl,
			StatusCode: response.StatusCode,
			Status:     response.Status,
		}
	}

	return response, nil
}
func (c *HttpClient) ResolveRedirect(ctx context.Context, sourceUrl *url.URL) (*url.URL, error) {
	redirectedUrl := sourceUrl

	c.client.CheckRedirect = func(request *http.Request, _ []*http.Request) error {
		redirectedUrl = request.URL

		return nil
	}

	_, err := c.Request(ctx, http.MethodGet, sourceUrl)
	if err != nil {
		return nil, err
	}

	c.client.CheckRedirect = nil

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

func (c *HttpClient) Download(ctx context.Context, sourceUrl *url.URL) ([]byte, error) {
	response, err := c.Request(ctx, http.MethodGet, sourceUrl)
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

	return responseBody, nil
}

func UnmarshalJson(response *http.Response, value interface{}) error {
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := response.Body.Close(); err != nil {
		return err
	}

	return json.Unmarshal(responseBody, value)
}

type ErrHttpNotOk struct {
	Method     string
	Url        url.URL
	Status     string
	StatusCode int
}

func (e ErrHttpNotOk) Error() string {
	return fmt.Sprintf("%s \"%s\" returned with status code %s", e.Method, e.Url.String(), e.Status)
}
