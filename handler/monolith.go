package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/network"
	"os/exec"
)

var ErrMonolith = errors.New("monolith error")

type MonolithHandler struct {
	path string
	args []string
}

func NewMonolithHandler(cfg config.Config) DownloadHandler {
	if !cfg.File.Monolith.Enabled {
		return &NoOpHandler{}
	}

	args := []string{"--user-agent", cfg.File.Archive.UserAgent}

	switch {
	case cfg.File.Monolith.AllowInsecure:
		args = append(args, "--insecure")
		fallthrough
	case !cfg.File.Monolith.IncludeAudio:
		args = append(args, "--no-audio")
		fallthrough
	case !cfg.File.Monolith.IncludeCss:
		args = append(args, "--no-css")
		fallthrough
	case !cfg.File.Monolith.IncludeFonts:
		args = append(args, "--no-fonts")
		fallthrough
	case !cfg.File.Monolith.IncludeFrames:
		args = append(args, "--no-frames")
		fallthrough
	case !cfg.File.Monolith.IncludeImages:
		args = append(args, "--no-images")
		fallthrough
	case !cfg.File.Monolith.IncludeJs:
		args = append(args, "--no-js")
		fallthrough
	case !cfg.File.Monolith.IncludeVideo:
		args = append(args, "--no-video")
		fallthrough
	case !cfg.File.Monolith.IncludeMetadata:
		args = append(args, "--no-metadata")
	}

	return &MonolithHandler{path: cfg.File.Monolith.Path, args: args}
}

func (s *MonolithHandler) Handle(_ context.Context, response DownloadResponse) (SourceContent, error) {
	if response.MediaType() != network.HtmlMediaType {
		return SourceContent{}, ErrNotHandled
	}

	if _, err := exec.LookPath(s.path); err != nil {
		return SourceContent{}, ErrNotHandled
	}

	var args []string
	args = append(args, s.args...)
	args = append(args, "--base-url", response.Url.String(), "-")

	command := exec.Command(s.path, args...)
	command.Stdin = bytes.NewReader(response.Body)

	stdout, err := command.Output()
	if err != nil {
		return SourceContent{}, fmt.Errorf("%w: %v", ErrMonolith, err)
	}

	return SourceContent{
		Content:   stdout,
		MediaType: response.MediaType(),
		FileName:  config.InferFileName(&response.Url, response.MediaType(), response.Header),
	}, nil
}
