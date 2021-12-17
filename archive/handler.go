package archive

import "io"

type HandlerFactory interface {
	Handler(mediaType string) Handler
}

type Handler interface {
	Handle(content io.ReadCloser) (io.ReadCloser, error)
}

type PassThroughHandler struct{}

func (PassThroughHandler) Handle(reader io.ReadCloser) (io.ReadCloser, error) {
	return reader, nil
}
