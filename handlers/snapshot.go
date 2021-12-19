package handlers

import (
	"bytes"
	"context"
	"os/exec"
)

type HtmlSnapshotHandler struct {
	path string
}

func NewHtmlSnapshotHandler() *HtmlSnapshotHandler {
	return &HtmlSnapshotHandler{"monolith"}
}

func (s *HtmlSnapshotHandler) Handle(_ context.Context, response *HttpResponse) (*SourceContent, error) {
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
