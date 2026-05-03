package main

import (
	"dos/internal/common/connect"
	"dos/internal/services/client/domain"
	"dos/internal/services/client/transport"
	"fmt"
)

type App struct {
	Config  *Config
	Conn    *connect.ConnCache
	Master  *transport.MasterTransport
	Storage *transport.StorageTransport
	Service *domain.Service
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

	storage, err := transport.NewStorageTransport(conn, &cfg.Storage)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init storage transport: %w", err)
	}

	service, err := domain.New(master, storage)
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
	}
	return app, nil
}
