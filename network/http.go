package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const (
	DefaultTimeout  time.Duration = 1000 * 1000 * 1000 * 15
	UserAgentHeader string        = "User-Agent"
)

var defaultClient http.Client

var (
	ErrInvalidApiUrl     = errors.New("invalid API url")
	ErrUnmarshalResponse = errors.New("error unmarshalling HTTP response")
)

func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	defaultClient = http.Client{
		Timeout: DefaultTimeout,
		Jar:     jar,
	}
}

func responseIsOk(status int) bool {
	return status >= 200 && status < 300
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

func (c *HttpClient) Request(ctx context.Context, method string, requestUrl url.URL) (*http.Response, error) {
	return c.RequestWithHeaders(ctx, method, requestUrl, nil)
}

func (c *HttpClient) RequestWithHeaders(ctx context.Context, method string, requestUrl url.URL, headers map[string]string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHttp, err)
	}

	request.Header.Set(UserAgentHeader, c.userAgent)

	for headerName, headerValue := range headers {
		request.Header.Set(headerName, headerValue)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHttp, err)
	}

	if !responseIsOk(response.StatusCode) {
		return nil, &HttpStatusError{
			Method:     method,
			Url:        requestUrl,
			StatusCode: response.StatusCode,
			Status:     response.Status,
		}
	}

	return response, nil
}
func (c *HttpClient) ResolveRedirect(ctx context.Context, sourceUrl url.URL) (url.URL, error) {
	redirectedUrl := sourceUrl

	c.client.CheckRedirect = func(request *http.Request, _ []*http.Request) error {
		redirectedUrl = *request.URL

		return nil
	}

	response, err := c.Request(ctx, http.MethodGet, sourceUrl)
	if err != nil {
		return url.URL{}, err
	}

	if err := response.Body.Close(); err != nil {
		return url.URL{}, err
	}

	c.client.CheckRedirect = nil

	return redirectedUrl, nil
}

func (c *HttpClient) CheckExists(ctx context.Context, sourceUrl url.URL) (bool, error) {
	response, err := c.Request(ctx, http.MethodGet, sourceUrl)

	defer func() {
		if response != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				err = closeErr
			}
		}
	}()

	if err != nil {
		httpErr := HttpStatusError{}
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			err = nil
		}

		return false, err
	}

	return true, err
}

func UnmarshalJson(response *http.Response, value interface{}) error {
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrHttp, err)
	}

	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrHttp, err)
	}

	if err := json.Unmarshal(responseBody, value); err != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshalResponse, err)
	}

	return nil
}

var ErrHttp = errors.New("http error")

type HttpStatusError struct {
	Method     string
	Url        url.URL
	Status     string
	StatusCode int
}

func (e HttpStatusError) Error() string {
	return fmt.Sprintf("%s \"%s\" returned %s", e.Method, e.Url.String(), e.Status)
}

func (e HttpStatusError) Unwrap() error {
	return ErrHttp
}
