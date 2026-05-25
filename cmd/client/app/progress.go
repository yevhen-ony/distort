package app 

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gosuri/uilive"
)

type ProgressConfig interface {
	RenderRefreshInterval() time.Duration
}

type ProgressRender struct {
	out      *uilive.Writer
	interval time.Duration
	state fmt.Stringer
	stop chan struct{}
	mu sync.Mutex
}

func NewProgressRender(refreshInterval time.Duration) *ProgressRender {
	return &ProgressRender{
		out: uilive.New(),
		interval: refreshInterval,
		stop: make(chan struct{}, 1),
	}
}

func (r *ProgressRender) Close() error {
	r.stop <- struct{}{}
	return nil
}

func (r *ProgressRender) Update(state fmt.Stringer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state = state
}

func (r *ProgressRender) renderState()  {

	r.mu.Lock()
	if r.state == nil {
		r.mu.Unlock()
		return
	}

	str := r.state.String()
	r.mu.Unlock()

	_, _ = fmt.Fprintln(r.out, str)
}

func (r *ProgressRender) RunLoop(ctx context.Context) {

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.out.Start()
	defer r.out.Stop()

	for {
		select {
		case <-r.stop:
			r.renderState()		
			return 
		case <-ctx.Done():
			r.renderState()
			return	
		case <-ticker.C:
			r.renderState()		
		}
	}
}
