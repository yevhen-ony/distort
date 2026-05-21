package replicate

import "dos/internal/common/metrics"

type ExecutorMetrics struct {
	UnreachableChunkObservationsTotal metrics.Counter
	ReplicationScheduledTotal         metrics.Counter
	DeleteReplicaSuccessDuration      metrics.Histogram
	DeleteReplicaFailedDuration       metrics.Histogram
	AddReplicaSuccessDuration         metrics.Histogram
	AddReplicaFailedDuration          metrics.Histogram
}


func NewExecutorMetrics(provider metrics.Provider) *ExecutorMetrics {
	buckets := []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}

  	unreachableChunkObservationsTotal := provider.Counter(metrics.CounterOpts{
  		Name: "master_replication_unreachable_chunk_observations_total",
  		Help: "Total number of times replication encountered a chunk with zero available replicas.",
  	})

  	replicationScheduledTotal := provider.Counter(metrics.CounterOpts{
  		Name: "master_replication_scheduled_total",
  		Help: "Total number of chunk replication tasks scheduled.",
  	})

  	deleteReplicaSuccessDuration := provider.Histogram(metrics.HistogramOpts{
  		Name: "master_replication_delete_replica_success_duration_seconds",
  		Help: "Duration of successful replica delete attempts in the replication executor.",
		Buckets: buckets,
  	})

  	deleteReplicaFailedDuration := provider.Histogram(metrics.HistogramOpts{
  		Name: "master_replication_delete_replica_failed_duration_seconds",
  		Help: "Duration of failed replica delete attempts in the replication executor.",
		Buckets: buckets,
  	})

  	addReplicaSuccessDuration := provider.Histogram(metrics.HistogramOpts{
  		Name: "master_replication_add_replica_success_duration_seconds",
  		Help: "Duration of successful replica add attempts in the replication executor.",
		Buckets: buckets,
  	})

  	addReplicaFailedDuration := provider.Histogram(metrics.HistogramOpts{
  		Name: "master_replication_add_replica_failed_duration_seconds",
  		Help: "Duration of failed replica add attempts in the replication executor.",
		Buckets: buckets,
  	})

  	return &ExecutorMetrics{
  		UnreachableChunkObservationsTotal: unreachableChunkObservationsTotal,
  		ReplicationScheduledTotal:         replicationScheduledTotal,
  		DeleteReplicaSuccessDuration:      deleteReplicaSuccessDuration,
  		DeleteReplicaFailedDuration:       deleteReplicaFailedDuration,
  		AddReplicaSuccessDuration:         addReplicaSuccessDuration,
  		AddReplicaFailedDuration:          addReplicaFailedDuration,
  	}
}

