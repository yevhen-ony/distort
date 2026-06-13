package heartbeat

import "dos/internal/common/metrics"


type HeartbeatMetrics struct {
	HeartbeatFailedTotal metrics.Counter
}

func NewHeartbeatMetrics(provider metrics.Provider) *HeartbeatMetrics {
	heartbeatFailedTotal := provider.Counter(metrics.CounterOpts{
		Name: "storage_heartbeat_failed_total",
		Help: "Total number of failed heartbeat requests from storage to master.",
	})

	return &HeartbeatMetrics{
		HeartbeatFailedTotal: heartbeatFailedTotal,
	}
}

