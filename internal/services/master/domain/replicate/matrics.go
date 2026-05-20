package replicate

import "dos/internal/common/metrics"

type ReplicationMetrics struct {
	UnreachableChunkObservationsTotal metrics.Counter
	ReplicationScheduledTotal         metrics.Counter
	DeleteReplicaSuccessDuration      metrics.Histogram
	DeleteReplicaFailedDuration       metrics.Histogram
	AddReplicaSuccessDuration         metrics.Histogram
	AddReplicaFailedDuration          metrics.Histogram
}
