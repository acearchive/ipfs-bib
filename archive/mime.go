package archive

import (
	"mime"
)

type ContentDisposition string

const (
	ContentDispositionHeader = "Content-Disposition"
	DefaultFileName          = "source"
	DefaultMediaType         = "application/octet-stream"
)

func defaultFileName(mediaType string) (string, error) {
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		return "", err
	}

	if extensions == nil {
		return DefaultFileName, nil
	} else {
		return DefaultFileName + extensions[0], nil
	}
}

func GetFileName(disposition, contentType string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = DefaultMediaType
	}

	if disposition == "" {
		return defaultFileName(mediaType)
	} else {
		contentDisposition, dispositionParams, err := mime.ParseMediaType(disposition)
		if err != nil {
			return "", err
		}

		if dispositionFilename, ok := dispositionParams["filename"]; contentDisposition == "attachment" && ok {
			return dispositionFilename, nil
		} else {
			return defaultFileName(mediaType)
		}
	}
}
