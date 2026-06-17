package report

import (
	"context"
	"errors"
	"testing"
	"time"

	"dos/internal/common/metrics"
	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestReportService_RunReportIteration_RetriesPendingReports(tt *testing.T) {
	ctx := context.Background()
	f := newReportFixutre(tt)

	f.config.EXPECT().QueueCapacity().Return(10)
	f.config.EXPECT().ReportInterval().Return(time.Hour)

	report := t.NewReplicaDeleted("chunk-1").ToRecord()
	reports := []t.StorageNodeReport{report}
	result := t.ReportResult{Accepted: []t.ChunkID{"chunk-1"}}
	sendErr := errors.New("send failed")

	gomock.InOrder(
		// first report iteration, use enqueued reports
		f.identity.EXPECT().
			GetID().
			Return(t.NodeID("node-1"), nil),
		f.transport.EXPECT().
			ReportChunks(ctx, t.NodeID("node-1"), reports).
			Return(t.ReportResult{}, sendErr),

		// second report iteration, use pending reports
		f.identity.EXPECT().
			GetID().
			Return(t.NodeID("node-1"), nil),
		f.transport.EXPECT().
			ReportChunks(ctx, t.NodeID("node-1"), reports).
			Return(result, nil),

		// processing
		f.processor.EXPECT().
			ProcessReport(ctx, result),
	)

	s, err := NewReportService(f.deps())
	require.NoError(tt, err)
	s.SetReportProcessor(f.processor)

	// do report
	s.Report(ctx, report)

	// first operation, expect to fail -> see mock setup
	s.RunReportIteration(ctx)
	require.Equal(tt, reports, s.pending)

	// second operation, expect to succeed
	s.RunReportIteration(ctx)
	require.Empty(tt, s.pending)
}

// fixture

type reportFixture struct {
	identity  *MockIdentityProvider
	transport *MockMasterTransport
	config    *MockReportConfig
	processor *MockReportProcessor
}

func newReportFixutre(tt *testing.T) *reportFixture {
	ctrl := gomock.NewController(tt)

	return &reportFixture{
		identity:  NewMockIdentityProvider(ctrl),
		transport: NewMockMasterTransport(ctrl),
		config:    NewMockReportConfig(ctrl),
		processor: NewMockReportProcessor(ctrl),
	}
}

func (f *reportFixture) deps() ReportDeps {
	return ReportDeps{
		Identity: f.identity,
		MasterT:  f.transport,
		Config:   f.config,
		Metrics:  NewReportMetrics(metrics.NopProvider{}),
	}
}
