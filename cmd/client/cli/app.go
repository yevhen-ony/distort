package main

import (
	"fmt"

	"github.com/gosuri/uilive"

	"dos/internal/common/connect"
	"dos/internal/common/transport/chunkrpc"
	"dos/internal/services/client/domain"
	"dos/internal/services/client/transport"
)

type App struct {
	Config  *Config
	Conn    *connect.ConnCache
	Master  *transport.MasterTransport
	Storage *chunkrpc.Transport
	Service *domain.Service

	progressOutput *uilive.Writer
}

func (app *App) Close() error {
	if app == nil || app.Conn == nil {
		return nil
	}
	return app.Conn.Close()
}

func NewApp(cfg *Config) (*App, error) {
	conn := connect.NewConnCache()

	master, err := transport.NewMasterTransport(conn, &cfg.Master)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init master transport: %w", err)
	}

	storage, err := chunkrpc.NewTransport(conn, &cfg.Storage)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init storage transport: %w", err)
	}

	output := uilive.New()
	opt := domain.WithProgressHandler(func(op *domain.ObjectProgress) {
		fmt.Fprint(output, op.String())
	 	_ = output.Flush()
	})

	service, err := domain.NewService(master, storage, opt)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init service: %w", err)
	}

	app := &App{
		Config:  cfg,
		Conn:    conn,
		Master:  master,
		Storage: storage,
		Service: service,

		progressOutput: output,
	}
	return app, nil
}
