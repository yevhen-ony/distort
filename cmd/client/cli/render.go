package main

import (
	"dos/cmd/client/app"
	"dos/cmd/client/app/render"
	"errors"
	"os"
)

type Render interface {
	Ping(ping *app.PingResult) error
	Error(string, error) error
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
