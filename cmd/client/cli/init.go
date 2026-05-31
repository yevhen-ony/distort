package main

import (
	"dos/cmd/client/app"
	"fmt"
)

type AppHolder struct {
	Config *app.Config
	App    *app.App
	Render Render
}

func InitApp(config *app.Config) (*AppHolder, error) {
	var err error
	a := &AppHolder{Config: config}

	a.App, err = app.NewApp(config)
	if err != nil {
		return nil, fmt.Errorf("init app: %w", err)
	}

	a.Render, err = NewRender(&config.CLI)
	if err != nil {
		return nil, fmt.Errorf("init render: %w", err)
	}
	
	return a, nil
}

func (a *AppHolder) Close() error {
	return a.App.Close()
}
