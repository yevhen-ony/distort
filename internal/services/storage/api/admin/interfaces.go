package admin

import (
	"context"

	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

//go:generate mockgen -source=$GOFILE -destination=mocks_test.go -package=admin

type Inventory interface {
	GetStats() t.NodeStats
	ListRecords() []s.ChunkRecord
}

type Storage interface {
	StageAndReportMany(context.Context, []t.ChunkID) *s.TriggerReportResult
	StageAndReportAll(context.Context) *s.TriggerReportResult
}

type Heartbeat interface {
	Pause() bool
	Resume() bool
	IsPaused() bool
}
