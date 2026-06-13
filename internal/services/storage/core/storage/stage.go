package storage

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
	"log/slog"
)

func (cs *StorageService) StageAndReportOne(
	ctx context.Context,
	chunkID t.ChunkID,
) error {
	meta, err := cs.inventory.Stage(chunkID)
	if err != nil {
		return err
	}
	cs.reporter.Report(ctx, t.NewReplicaStaged(meta).ToRecord())
	return nil
}

func (cs *StorageService) StageAndReportMany(
	ctx context.Context,
	chunkIDs []t.ChunkID,
) *s.TriggerReportResult {

	res := s.NewTriggerReportResult()

	for _, chunkID := range chunkIDs {
		if err := cs.StageAndReportOne(ctx, chunkID); err != nil {
			res.Failed = append(res.Failed, chunkID)
			continue
		}
		res.Scheduled = append(res.Scheduled, chunkID)
	}
	cs.reporter.Flush(ctx)

	return res
}

func (cs *StorageService) StageAndReportAll(ctx context.Context) *s.TriggerReportResult {

	slog.InfoContext(ctx, "stage and report all chunks in catalog")
	return cs.StageAndReportMany(ctx, cs.inventory.ListIDs())
}

func (cs *StorageService) ProcessReport(ctx context.Context, r t.ReportResult) {
	ctx = dosctx.WithOperation(ctx, "process_report")

	for _, chunkID := range r.Accepted {
		if _, err := cs.inventory.Activate(chunkID); err != nil {
			slog.WarnContext(ctx, "activation failed", "chunk_id", chunkID, "error", err)
		}
	}

	for _, chunkID := range r.Rejected {
		rec, err := cs.inventory.GetRecord(chunkID)
		if err == nil && rec.State == s.ChunkStateStaged {
			cs.reporter.Report(ctx, t.NewReplicaStaged(rec.Meta).ToRecord())
		}
	}
}
