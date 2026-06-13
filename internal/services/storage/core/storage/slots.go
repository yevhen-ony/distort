package storage

import (
	"context"
	"time"

	s "dos/internal/services/storage"
)

func (cs *StorageService) AcquireOpSlot(ctx context.Context) (func(), error) {
	acqCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	start := time.Now()

	select {
	case <-acqCtx.Done():
		cs.metrics.OpSlotsAcquireDuration.Observe(time.Since(start).Seconds())
		return nil, s.ErrServiceBusy

	case cs.sem <- struct{}{}:
		cs.metrics.OpSlotsAcquireDuration.Observe(time.Since(start).Seconds())
		cs.metrics.OpSlotsInUse.Add(1)
		start = time.Now()
		release := func() {
			<-cs.sem
			cs.metrics.OpSlotsInUse.Add(-1)
			cs.metrics.OpSlotsHoldDuration.Observe(time.Since(start).Seconds())
		}
		return release, nil
	}
}
