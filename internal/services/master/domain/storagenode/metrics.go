package storagenode

import "dos/internal/common/metrics"

type NodeReportMetrics struct {
	StagedReplicasAcceptedTotal	metrics.Counter
	StagedReplicasRejectedTotal metrics.Counter
	DeletedReplicasTotal metrics.Counter
	ReplicaChainFailedTotal metrics.Counter
}

type NodeLifecycleMetrics struct {
	RegisteredNodesCount metrics.Gauge
}
