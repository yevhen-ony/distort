package main

import (
	"dos/cmd/client/app"
	"dos/cmd/client/app/render"
	"errors"
	"os"
)

type Render interface {

	Error(string, error) error

	Ping(*app.PingResult) error

	ListObjects(*app.ListObjectsResult) error
	ListChunks(*app.ListChunksResult) error
	ListNodes(*app.ListNodesResult) error
	
	DiscoverMaster(*app.DiscoverMasterResult) error
}

func NewRender(config *app.CLIConfig) (Render, error) {
	switch config.OutputFormat {
	case "json":
		return render.NewJSONRender(os.Stdout, true)
	case "text":
		return render.NewTextRender(os.Stdout)
	default:
		return nil, errors.New("unknown output format")
	}
}
