package report

import "dos/internal/common/metrics"

type ReportServiceMetrics struct {
	ReportsSentTotal      metrics.Counter
	ReportsFailedTotal    metrics.Counter
	ReportsRecordsTotal   metrics.Counter
	ReportsRejectedTotal  metrics.Counter
	ReportsQueueBatchSize metrics.Histogram
}
