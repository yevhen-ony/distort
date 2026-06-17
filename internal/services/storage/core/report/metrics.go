package report

import "dos/internal/common/metrics"

type ReportMetrics struct {
	ReportsSentTotal      metrics.Counter
	ReportsFailedTotal    metrics.Counter
	ReportsRecordsTotal   metrics.Counter
	ReportsRejectedTotal  metrics.Counter
	ReportsQueueBatchSize metrics.Histogram
}

func NewReportMetrics(provider metrics.Provider) *ReportMetrics {
	reportsSentTotal := provider.Counter(metrics.CounterOpts{
		Name: "storage_reports_sent_total",
		Help: "Total number of report requests sent from storage to master.",
	})

	reportsFailedTotal := provider.Counter(metrics.CounterOpts{
		Name: "storage_reports_failed_total",
		Help: "Total number of report requests that failed to send from storage to master.",
	})

	reportsRecordsTotal := provider.Counter(metrics.CounterOpts{
		Name: "storage_report_records_total",
		Help: "Total number of report records successfully delivered from storage to master.",
	})

	reportsRejectedTotal := provider.Counter(metrics.CounterOpts{
		Name: "storage_report_records_rejected_total",
		Help: "Total number of report records rejected by the master.",
	})

	reportsQueueBatchSize := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_reports_queue_batch_size",
		Help:    "Number of report records drained from the storage report queue per send iteration.",
		Buckets: []float64{1, 2, 4, 8, 16, 32, 64},
	})

	return &ReportMetrics{
		ReportsSentTotal:      reportsSentTotal,
		ReportsFailedTotal:    reportsFailedTotal,
		ReportsRecordsTotal:   reportsRecordsTotal,
		ReportsRejectedTotal:  reportsRejectedTotal,
		ReportsQueueBatchSize: reportsQueueBatchSize,
	}
}
