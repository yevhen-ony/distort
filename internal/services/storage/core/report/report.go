package report

import (
	"context"
	"dos/internal/common/dosctx"
	"dos/internal/common/loop"
	"dos/internal/common/queue"
	t "dos/internal/common/types"
	"dos/internal/services/storage/transport"
	"log/slog"
	"time"
)


type Config interface {
	ReportInterval() time.Duration
	QueueCapacity() int
}

type IdentityProvider interface {
	GetID() (t.NodeID, error)
}

type ReportProcessor interface{
	Process(context.Context, t.ReportResult)
}

type NOPReportProcessor struct {}
func (*NOPReportProcessor) Process(context.Context, t.ReportResult) {}

type ReportService struct {
	identity IdentityProvider 
	master *transport.Master

	config Config

	pending []t.StorageNodeReport
	processor ReportProcessor
	
	queue *queue.Queue[t.StorageNodeReport]
	looper *loop.Looper
}

func NewReportService(
	identity IdentityProvider, master *transport.Master, config Config, 
) *ReportService {
	queue := queue.NewQueue[t.StorageNodeReport](config.QueueCapacity())
	return &ReportService {
		identity: identity,
		master: master,
		config: config,
		queue: queue,
		processor: &NOPReportProcessor{},
		looper: loop.NewLooper(config.ReportInterval()),
	}
}

func (rs *ReportService) SetReportProcessor(p ReportProcessor) {
	rs.processor = p
}

func (rs *ReportService) Report(ctx context.Context, report t.StorageNodeReport) {
	if err := rs.queue.Enq(ctx, report); err != nil {
		return 
	}
	if rs.queue.Full() {
		rs.Flush(ctx)
	}
}


func (rs *ReportService) RunReportIteration(ctx context.Context) {

	slog.DebugContext(ctx, "exec report chunks")

	if len(rs.pending) > 0 {
		result, err := rs.SendReports(ctx, rs.pending)
		if err != nil {
			slog.ErrorContext(ctx, "report chunks failed", "error", err)
			return
		}
		rs.ProcessReportResult(ctx, result)
		rs.pending = nil
	}
	
	rs.pending = rs.queue.Drain()
	slog.DebugContext(ctx, "drain report queue", "length", len(rs.pending))
	if len(rs.pending) == 0 {
		return
	}

	result, err := rs.SendReports(ctx, rs.pending)
	if err != nil {
		slog.ErrorContext(ctx, "report chunks failed", "error", err)
		return
	}
	rs.ProcessReportResult(ctx, result)
	rs.pending = nil
}

func (rs *ReportService) ProcessReportResult(ctx context.Context, result t.ReportResult) {
	if len(result.Rejected) != 0 {
		slog.WarnContext(ctx, "rejected chunks", "chunk_ids", result.Rejected)
	}
	rs.processor.Process(ctx, result)
}

func (rs *ReportService) SendReports(
	ctx context.Context, reports []t.StorageNodeReport,
) (t.ReportResult, error) {
	nodeID, err := rs.identity.GetID()
	if err != nil {
		return t.ReportResult{}, err
	}

	result, err := rs.master.ReportChunks(ctx, nodeID, reports)
	if err != nil {
		return t.ReportResult{}, err
	}
	return result, nil
}

func (rs *ReportService) RunLoop(ctx context.Context) {
	ctx = dosctx.WithService(ctx, "reporter")
	rs.looper.SkipFirstWait().Run(ctx, rs.RunReportIteration) 
}

func (rs *ReportService) Flush(_ context.Context) {
	rs.looper.Flush()
}
