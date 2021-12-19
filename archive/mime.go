package archive

import (
	"mime"
)

type ContentDisposition string

const DefaultFileName = "source"

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

func InferFileName(disposition, mediaType string) (string, error) {
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
