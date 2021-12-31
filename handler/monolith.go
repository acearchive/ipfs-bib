package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"os/exec"
)

var ErrMonolith = errors.New("monolith error")

type MonolithHandler struct {
	path string
	args []string
}

func NewMonolithHandler(cfg config.Config) DownloadHandler {
	if !cfg.File.Snapshot.Enabled {
		return &NoOpHandler{}
	}

	args := []string{"--user-agent", cfg.File.Archive.UserAgent}

	switch {
	case cfg.File.Snapshot.AllowInsecure:
		args = append(args, "--insecure")
		fallthrough
	case !cfg.File.Snapshot.IncludeAudio:
		args = append(args, "--no-audio")
		fallthrough
	case !cfg.File.Snapshot.IncludeCss:
		args = append(args, "--no-css")
		fallthrough
	case !cfg.File.Snapshot.IncludeFonts:
		args = append(args, "--no-fonts")
		fallthrough
	case !cfg.File.Snapshot.IncludeFrames:
		args = append(args, "--no-frames")
		fallthrough
	case !cfg.File.Snapshot.IncludeImages:
		args = append(args, "--no-images")
		fallthrough
	case !cfg.File.Snapshot.IncludeJs:
		args = append(args, "--no-js")
		fallthrough
	case !cfg.File.Snapshot.IncludeVideo:
		args = append(args, "--no-video")
		fallthrough
	case !cfg.File.Snapshot.IncludeMetadata:
		args = append(args, "--no-metadata")
	}

	return &MonolithHandler{path: cfg.File.Snapshot.Path, args: args}
}

func (s *MonolithHandler) Handle(_ context.Context, response DownloadResponse) (SourceContent, error) {
	if response.MediaType() != "text/html" {
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
