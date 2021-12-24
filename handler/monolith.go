package handler

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

func NewMonolithHandler(cfg *config.Config) DownloadHandler {
	if !cfg.Snapshot.Enabled {
		return &NoOpHandler{}
	}

	args := []string{"--user-agent", cfg.Archive.UserAgent}

	switch {
	case cfg.Snapshot.AllowInsecure:
		args = append(args, "--insecure")
		fallthrough
	case !cfg.Snapshot.IncludeAudio:
		args = append(args, "--no-audio")
		fallthrough
	case !cfg.Snapshot.IncludeCss:
		args = append(args, "--no-css")
		fallthrough
	case !cfg.Snapshot.IncludeFonts:
		args = append(args, "--no-fonts")
		fallthrough
	case !cfg.Snapshot.IncludeFrames:
		args = append(args, "--no-frames")
		fallthrough
	case !cfg.Snapshot.IncludeImages:
		args = append(args, "--no-images")
		fallthrough
	case !cfg.Snapshot.IncludeJs:
		args = append(args, "--no-js")
		fallthrough
	case !cfg.Snapshot.IncludeVideo:
		args = append(args, "--no-video")
		fallthrough
	case !cfg.Snapshot.IncludeMetadata:
		args = append(args, "--no-metadata")
	}

	return &MonolithHandler{path: cfg.Snapshot.Path, args: args}
}

func (s *MonolithHandler) Handle(_ context.Context, response *DownloadResponse) (*SourceContent, error) {
	if response.MediaType() != "text/html" {
		return nil, ErrNotHandled
	}

	if _, err := exec.LookPath(s.path); err != nil {
		return nil, err
	}

	args := make([]string, len(s.args))
	copy(args, s.args)

	args = append(args, "--base-url", response.Url.String(), "-")

	command := exec.Command(s.path, args...)
	command.Stdin = bytes.NewReader(response.Body)

	stdout, err := command.Output()
	if err != nil {
		return nil, err
	}

	return &SourceContent{
		Content:   stdout,
		MediaType: response.MediaType(),
		FileName:  config.FileNameFromUrl(&response.Url, response.MediaType()),
	}, nil
}
