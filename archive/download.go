package archive

import (
	"io"
	"mime"
	"net/http"
	"net/url"
)

const (
	ContentTypeHeader        = "Content-Type"
	ContentDispositionHeader = "Content-Disposition"
	DefaultFilename          = "source"
)

func ParseMediaType(header http.Header) (medaType, filename string, err error) {
	mediaType, _, err := mime.ParseMediaType(header.Get(ContentTypeHeader))
	if err != nil {
		return "", "", err
	}

	contentDisposition, dispositionParams, err := mime.ParseMediaType(header.Get(ContentDispositionHeader))
	if err != nil {
		return "", "", err
	}

	if dispositionFilename, ok := dispositionParams["filename"]; contentDisposition == "attachment" && ok {
		filename = dispositionFilename
	} else {
		extensions, err := mime.ExtensionsByType(mediaType)
		if err != nil {
			return "", "", err
		}

		if extensions == nil {
			filename = DefaultFilename
		} else {
			filename = DefaultFilename + extensions[0]
		}
	}

	return mediaType, filename, err
}

func Download(url *url.URL, handlerFactory HandlerFactory) (content io.ReadCloser, filename string, err error) {
	response, err := http.Get(url.String())
	if err != nil {
		return nil, "", err
	}

	mediaType, filename, err := ParseMediaType(response.Header)
	if err != nil {
		return nil, "", err
	}

	handler := handlerFactory.Handler(mediaType)

	content, err = handler.Handle(response.Body)
	if err != nil {
		return nil, "", err
	}

	return content, filename, err
}
