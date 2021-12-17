package archive

import (
	shell "github.com/ipfs/go-ipfs-api"
	"net/http"
	"net/url"
)

const ContentTypeHeader = "Content-Type"

type Cid string

func AddToNode(node shell.Shell, handlerFactory HandlerFactory, pin bool, url *url.URL) (cid Cid, err error) {
	response, err := http.Get(url.String())
	if err != nil {
		return "", err
	}

	mediaType := response.Header.Get(ContentTypeHeader)
	handler := handlerFactory.Handler(mediaType)

	content, err := handler.Handle(response.Body)
	if err != nil {
		return "", err
	}

	defer func() {
		err = content.Close()
	}()

	rawCid, err := node.Add(content, shell.Pin(pin))
	if err != nil {
		return "", err
	}

	return Cid(rawCid), err
}
