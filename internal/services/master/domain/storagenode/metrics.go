package storagenode

import "dos/internal/common/metrics"

type ReportMetrics struct {
	StagedReplicasAcceptedTotal metrics.Counter
	StagedReplicasRejectedTotal metrics.Counter
	DeletedReplicasTotal        metrics.Counter
	ReplicaChainFailedTotal     metrics.Counter
}

type LifecycleMetrics struct {
	RegisteredNodesCount metrics.Gauge
}

func NewReportMetrics(provider metrics.Provider) *ReportMetrics {
	stagedReplicasAcceptedTotal := provider.Counter(metrics.CounterOpts{
		Name: "master_report_staged_replicas_accepted_total",
		Help: "Total number of staged replica reports accepted by the master.",
	})
	stagedReplicasRejectedTotal := provider.Counter(metrics.CounterOpts{
		Name: "master_report_staged_replicas_rejected_total",
		Help: "Total number of staged replica reports rejected by the master.",
	})
	deletedReplicasTotal := provider.Counter(metrics.CounterOpts{
		Name: "master_report_deleted_replicas_total",
		Help: "Total number of deleted replica reports processed by the master.",
	})
	replicaChainFailedTotal := provider.Counter(metrics.CounterOpts{
		Name: "master_report_replica_chain_failed_total",
		Help: "Total number of replica chain failure reports processed by the master.",
	})

	return &ReportMetrics{
		StagedReplicasAcceptedTotal: stagedReplicasAcceptedTotal,
		StagedReplicasRejectedTotal: stagedReplicasRejectedTotal,
		DeletedReplicasTotal:        deletedReplicasTotal,
		ReplicaChainFailedTotal:     replicaChainFailedTotal,
	}
}

func NewLifecycleMetrics(provider metrics.Provider) *LifecycleMetrics {
	registeredNodesCount := provider.Gauge(metrics.GaugeOpts{
		Name: "master_lifecycle_registered_nodes_count",
		Help: "Current number of storage nodes registered in the master node registry.",
	})

	return &LifecycleMetrics{
		RegisteredNodesCount: registeredNodesCount,
	}
}
