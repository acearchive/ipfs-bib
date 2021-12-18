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
	archiver *obelisk.Archiver
}

func NewDownloader() Downloader {
	timeout, err := time.ParseDuration(RequestTimeout)
	if err != nil {
		panic(err)
	}

	archiver := obelisk.Archiver{
		UserAgent:      obelisk.DefaultUserAgent,
		RequestTimeout: timeout,
	}

	archiver.Validate()

	return Downloader{&archiver}
}

func (a Downloader) Download(ctx context.Context, url *url.URL) (content []byte, filename string, err error) {
	// TODO: Don't perform two GET requests
	response, err := http.Get(url.String())
	if err != nil {
		return nil, "", err
	}

	content, contentType, err := a.archiver.Archive(ctx, obelisk.Request{
		URL: url.String(),
	})
	if err != nil {
		return nil, "", err
	}

	disposition := response.Header.Get(ContentDispositionHeader)

	filename, err = GetFileName(disposition, contentType)
	if err != nil {
		return nil, "", err
	}

	return content, filename, err
}
