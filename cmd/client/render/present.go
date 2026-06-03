package render

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"dos/cmd/client/app"
	"dos/internal/common/loop"
	"dos/internal/services/client/domain/progress"
)

type FrameFn func() ([]byte, error)

type WriteFlusher interface {
	io.Writer
	Flush() error
}

type Presenter struct {
	output WriteFlusher
	render Render

	mu      sync.Mutex
	current FrameFn
	looper  *loop.Looper
}

type PresenterDeps struct {
	Output   WriteFlusher
	Render   Render
	Interval time.Duration
}

func NewPresenter(deps PresenterDeps) (*Presenter, error) {
	if deps.Output == nil {
		return nil, errors.New("missing output")
	}
	if deps.Interval <= 0 {
		return nil, errors.New("missing interval")
	}
	if deps.Render == nil {
		return nil, errors.New("missing render")
	}
	p := &Presenter{
		output: deps.Output,
		render: deps.Render,
		looper: loop.NewLooper(deps.Interval),
	}
	return p, nil
}

func (p *Presenter) Flush() {
	p.looper.Flush()
}

func (p *Presenter) Present() error {
	p.mu.Lock()
	current := p.current
	p.mu.Unlock()

	if current == nil {
		return nil
	}

	b, err := current()
	if err != nil {
		return fmt.Errorf("get current: %w", err)
	}
	_, err = p.output.Write(b)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if err = p.output.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}

	return nil
}

func (p *Presenter) RunLoop(ctx context.Context) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	go p.looper.Run(ctx, func(context.Context) {
		_ = p.Present()
	})
	return cancel
}

func (p *Presenter) Update(state any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	current, err := p.resolve(state)
	if err != nil {
		return err
	}

	p.current = current
	return nil
}

func (p *Presenter) resolve(state any) (FrameFn, error) {
	switch s := state.(type) {

	case *ErrorResult:
		frameFn := func() ([]byte, error) { return p.render.Error(s) }
		return frameFn, nil

	case *app.PingResult:
		frameFn := func() ([]byte, error) { return p.render.Ping(s) }
		return frameFn, nil

	case *app.DiscoverMasterResult:
		frameFn := func() ([]byte, error) { return p.render.DiscoverMaster(s) }
		return frameFn, nil

	case *app.ListObjectsResult:
		frameFn := func() ([]byte, error) { return p.render.ListObjects(s) }
		return frameFn, nil

	case *app.ListChunksResult:
		frameFn := func() ([]byte, error) { return p.render.ListChunks(s) }
		return frameFn, nil

	case *app.ListNodesResult:
		frameFn := func() ([]byte, error) { return p.render.ListNodes(s) }
		return frameFn, nil

	case *app.DescribeChunkResult:
		frameFn := func() ([]byte, error) { return p.render.DescribeChunk(s) }
		return frameFn, nil

	case *app.DescribeObjectResult:
		frameFn := func() ([]byte, error) { return p.render.DescribeObject(s) }
		return frameFn, nil

	case *app.DownloadChunkResult:
		frameFn := func() ([]byte, error) { return p.render.DownloadChunk(s) }
		return frameFn, nil

	case *app.AllocateChunkResult:
		frameFn := func() ([]byte, error) { return p.render.AllocateChunk(s) }
		return frameFn, nil

	case *app.PushChunkResult:
		frameFn := func() ([]byte, error) { return p.render.PushChunk(s) }
		return frameFn, nil

	case *app.CreateObjectResult:
		frameFn := func() ([]byte, error) { return p.render.CreateObject(s) }
		return frameFn, nil
	
	case *app.InspectNodeResult:
		frameFn := func() ([]byte, error) { return p.render.InspectNode(s) }
		return frameFn, nil

  	case *app.TriggerReportResult:
		frameFn := func() ([]byte, error) { return p.render.TriggerReport(s) }
		return frameFn, nil

	case *progress.ObjectProgress:
		frameFn := func() ([]byte, error) { return p.render.Progress(s) }
		return frameFn, nil

	default:
		slog.Error(fmt.Sprintf("presenter: unsupported state type %T", state))
		return nil, errors.New("unsupported state")

	}
}
