package archive

import (
	"mime"
	"net/http"
	"net/url"
	"path"
)

type ContentDisposition string

const (
	ContentTypeHeader        = "Content-Type"
	ContentDispositionHeader = "Content-Disposition"
	DefaultMediaType         = "application/octet-stream"
)

type MediaInfo struct {
	MediaType string
	Filename  string
}

func fileNameFromUrl(url *url.URL) string {
	return path.Base(url.Path)
}

func parseMediaType(header http.Header) (string, error) {
	var (
		mediaType string
		err       error
	)

	contentTypeHeader := header.Get(ContentTypeHeader)

	if contentTypeHeader == "" {
		mediaType = DefaultMediaType
	} else {
		mediaType, _, err = mime.ParseMediaType(header.Get(ContentTypeHeader))
		if err != nil {
			return "", err
		}
	}

	return mediaType, err
}

func ParseMediaInfo(url *url.URL, header http.Header) (*MediaInfo, error) {
	mediaType, err := parseMediaType(header)
	if err != nil {
		return nil, err
	}

	var filename string

	dispositionHeader := header.Get(ContentDispositionHeader)

	if dispositionHeader == "" {
		filename = fileNameFromUrl(url)
	} else {
		contentDisposition, dispositionParams, err := mime.ParseMediaType(header.Get(ContentDispositionHeader))
		if err != nil {
			return nil, err
		}

		if dispositionFilename, ok := dispositionParams["filename"]; contentDisposition == "attachment" && ok {
			filename = dispositionFilename
		} else {
			extensions, err := mime.ExtensionsByType(mediaType)
			if err != nil {
				return nil, err
			}

			if extensions == nil {
				filename = fileNameFromUrl(url)
			} else {
				filename = fileNameFromUrl(url) + extensions[0]
			}
		}
	}

	return &MediaInfo{
		MediaType: mediaType,
		Filename:  filename,
	}, nil
}
