package archive

import (
	"context"
	"github.com/go-shiori/obelisk"
	"net/http"
	"net/url"
	"time"
)

const RequestTimeout = "30s"

type Downloader struct {
	archiver obelisk.Archiver
}

func NewDownloader() *Downloader {
	timeout, err := time.ParseDuration(RequestTimeout)
	if err != nil {
		panic(err)
	}

	return &Downloader{obelisk.Archiver{
		UserAgent:      obelisk.DefaultUserAgent,
		RequestTimeout: timeout,
	}}
}

func (a *Downloader) Download(ctx context.Context, url *url.URL) (content []byte, info *MediaInfo, err error) {
	response, err := http.Get(url.String())
	if err != nil {
		return nil, nil, err
	}

	info, err = ParseMediaType(response.Header)
	if err != nil {
		return nil, nil, err
	}

	content, _, err = a.archiver.Archive(ctx, obelisk.Request{
		Input: response.Body,
		URL:   url.String(),
	})
	if err != nil {
		return nil, nil, err
	}

	err = response.Body.Close()

	return content, info, err
}
