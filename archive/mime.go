package archive

import (
	"mime"
	"net/http"
)

type ContentDisposition string

const (
	InlineDisposition     ContentDisposition = "inline"
	AttachmentDisposition ContentDisposition = "attachment"

	ContentTypeHeader        string = "Content-Type"
	ContentDispositionHeader string = "Content-Disposition"
	DefaultFilename          string = "source"
)

type MediaInfo struct {
	Disposition ContentDisposition
	MediaType   string
	Filename    string
}

func ParseMediaType(header http.Header) (*MediaInfo, error) {
	mediaType, _, err := mime.ParseMediaType(header.Get(ContentTypeHeader))
	if err != nil {
		return nil, err
	}

	contentDisposition, dispositionParams, err := mime.ParseMediaType(header.Get(ContentDispositionHeader))
	if err != nil {
		return nil, err
	}

	disposition := ContentDisposition(contentDisposition)

	var filename string

	if dispositionFilename, ok := dispositionParams["filename"]; disposition == AttachmentDisposition && ok {
		filename = dispositionFilename
	} else {
		extensions, err := mime.ExtensionsByType(mediaType)
		if err != nil {
			return nil, err
		}

		if extensions == nil {
			filename = DefaultFilename
		} else {
			filename = DefaultFilename + extensions[0]
		}
	}

	return &MediaInfo{
		Disposition: disposition,
		MediaType:   mediaType,
		Filename:    filename,
	}, nil
}
