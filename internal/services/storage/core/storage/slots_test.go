package storage

import (
	"context"
	"testing"
	"time"

	"dos/internal/common/metrics"
	s "dos/internal/services/storage"

	"github.com/stretchr/testify/require"
)

func TestStorageService_AcquireOpSlot_AcquireAndRelease(tt *testing.T) {
	ctx := context.Background()

	service := &StorageService{
		metrics: NewStorageMetrics(metrics.NopProvider{}),
		sem:     make(chan struct{}, 1),
	}

	release, err := service.AcquireOpSlot(ctx, defaultOpSlotAcquireTimeout)
	require.NoError(tt, err)
	require.NotNil(tt, release)
	require.Len(tt, service.sem, 1)

	release()

	require.Len(tt, service.sem, 0)
}

func TestStorageService_AcquireOpSlot_OnFull(tt *testing.T) {
	ctx := context.Background()

	service := &StorageService{
		metrics: NewStorageMetrics(metrics.NopProvider{}),
		sem:     make(chan struct{}, 1),
	}

	release, err := service.AcquireOpSlot(ctx, time.Millisecond)
	require.NoError(tt, err)
	defer release()

	_, err = service.AcquireOpSlot(ctx, time.Millisecond)
	require.ErrorIs(tt, err, s.ErrServiceBusy)
}
