package handlers

import (
	"bytes"
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"os/exec"
)

type MonolithHandler struct {
	path string
}

func NewMonolithHandler(cfg *config.MonolithHandler) DownloadHandler {
	if cfg.Enabled {
		return &MonolithHandler{path: cfg.Path}
	} else {
		return &NoOpHandler{}
	}
}

func (s *MonolithHandler) Handle(_ context.Context, response *HttpResponse) (*SourceContent, error) {
	if response.MediaType() != "text/html" {
		return nil, nil
	}

	if _, err := exec.LookPath(s.path); err != nil {
		return nil, nil
	}

	command := exec.Command(s.path, "-")
	command.Stdin = bytes.NewReader(response.Body)

	stdout, err := command.Output()
	if err != nil {
		return nil, err
	}

	return &SourceContent{
		Content:   stdout,
		MediaType: response.MediaType(),
	}, nil
}
