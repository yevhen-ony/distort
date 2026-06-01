package main

import (
	"context"
	"fmt"
	"io"

	"dos/cmd/client/app"
	"dos/cmd/client/render"

	"github.com/gosuri/uilive"
)

type AppHolder struct {
	Out       *uilive.Writer
	Config    *app.Config
	App       *app.App
	Render    render.Render
	Presenter *render.Presenter

	cancel context.CancelFunc
}

func RunApp(ctx context.Context, config *app.Config) (*AppHolder, error) {

	var err error
	a := &AppHolder{Config: config}

	a.App, err = app.NewApp(config)
	if err != nil {
		return nil, fmt.Errorf("init app: %w", err)
	}

	a.Render, err = render.NewRender(config.OutputFormat())
	if err != nil {
		return nil, fmt.Errorf("init render: %w", err)
	}

	a.Out = uilive.New()
	a.Out.Start()

	a.Presenter, err = render.NewPresenter(render.PresenterDeps{
		Output:   a.Out,
		Render:   a.Render,
		Interval: config.RenderRefreshInterval(),
	})
	if err != nil {
		return nil, fmt.Errorf("init presenter: %w", err)
	}

	return a, nil
}

func (a *AppHolder) UpdateOutput(h func(io.Writer) error) {
	a.Presenter.Update(h)
}

func (a *AppHolder) Close() error {
	a.Presenter.Present()

	if a.cancel != nil {
		a.cancel()
	}
	a.Out.Stop()
	return a.App.Close()
}
