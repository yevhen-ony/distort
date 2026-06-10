package app

import (
	"errors"
	"fmt"

	"dos/internal/common/connect"
	"dos/internal/common/master/resolve"
	"dos/internal/common/master/route"
	"dos/internal/common/transport/chunkrpc"
	"dos/internal/common/transport/healthrpc"
	"dos/internal/common/transport/storage/adminrpc"
	"dos/internal/services/client/domain/progress"
	"dos/internal/services/client/transport"
)

type App struct {
	Config *Config

	Conn           *connect.ConnCache
	Master         *MasterHolder
	ChunkT         *chunkrpc.Transport
	StorageHealthT *healthrpc.Transport
	StorageAdminT  *adminrpc.Transport

	onProgress func(*progress.ObjectProgress)
}

func (app *App) SetOnProgress(fn func(*progress.ObjectProgress)) {
	app.onProgress = fn
}

func (app *App) Close() error {
	_ = app.Master.router.Close()
	return app.Conn.Close()
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	if config == nil {
		return nil, errors.New("missing config")
	}

	master, err := initMasterTransport(config)
	if err != nil {
		return nil, err
	}

	chunkT, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init storage transport: %w", err)
	}

	storageHealthT, err := healthrpc.NewTransport(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init health transport: %w", err)
	}

	storageAdminT, err := adminrpc.NewTransport(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init admin storage node transport: %w", err)
	}

	app := &App{
		Config:         config,
		Conn:           conn,
		Master:         master,
		ChunkT:         chunkT,
		StorageHealthT: storageHealthT,
		StorageAdminT:  storageAdminT,
	}
	return app, nil
}

func (app *App) MasterT() *transport.MasterTransport {
	return app.Master.transport
}

type MasterHolder struct {
	resolver  *resolve.Resolver
	router    *route.MasterRouter
	transport *transport.MasterTransport
}

func initMasterTransport(config *Config) (*MasterHolder, error) {
	mresolver, err := resolve.New(&config.Master)
	if err != nil {
		return nil, fmt.Errorf("master resolver init: %w", err)

	}
	mrouter, err := route.New(mresolver)
	if err != nil {
		return nil, fmt.Errorf("master router init: %w", err)
	}

	mtransport, err := transport.NewMasterTransport(transport.MasterTransportDeps{
		Router: mrouter,
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("master transport init: %w", err)
	}

	holder := &MasterHolder{
		resolver:  mresolver,
		router:    mrouter,
		transport: mtransport,
	}
	return holder, nil
}
