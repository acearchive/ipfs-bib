package handlers

import (
	"bytes"
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"os/exec"
)

type MonolithHandler struct {
	path string
	args []string
}

func NewMonolithHandler(cfg *config.MonolithHandler) DownloadHandler {
	if !cfg.Enabled {
		return &NoOpHandler{}
	}

	args := []string{"--output", "-"}

	switch {
	case cfg.AllowInsecure:
		args = append(args, "--insecure")
		fallthrough
	case !cfg.IncludeAudio:
		args = append(args, "--no-audio")
		fallthrough
	case !cfg.IncludeCss:
		args = append(args, "--no-css")
		fallthrough
	case !cfg.IncludeFonts:
		args = append(args, "--no-fonts")
		fallthrough
	case !cfg.IncludeFrames:
		args = append(args, "--no-frames")
		fallthrough
	case !cfg.IncludeImages:
		args = append(args, "--no-images")
		fallthrough
	case !cfg.IncludeJs:
		args = append(args, "--no-js")
		fallthrough
	case !cfg.IncludeVideo:
		args = append(args, "--no-video")
		fallthrough
	case !cfg.IncludeMetadata:
		args = append(args, "--no-metadata")
	}

	return &MonolithHandler{path: cfg.Path, args: args}
}

func (s *MonolithHandler) Handle(_ context.Context, response *HttpResponse) (*SourceContent, error) {
	if response.MediaType() != "text/html" {
		return nil, nil
	}

	if _, err := exec.LookPath(s.path); err != nil {
		return nil, nil
	}

	command := exec.Command(s.path, s.args...)
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
