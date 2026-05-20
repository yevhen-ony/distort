package catalog

import "dos/internal/common/metrics"

type CatalogMetrics struct {
	ObjectCount metrics.Gauge
	ChunkCount metrics.Gauge
}

func NewCatalogMetrics(provider metrics.Provider) *CatalogMetrics {
	objectCount := provider.Gauge(metrics.GaugeOpts{
		Name: "catalog_objects_count",
		Help: "Current number of objects in the master catalog.",
	})
	chunkCount := provider.Gauge(metrics.GaugeOpts{
		Name: "catalog_chunks_count",
		Help: "Current number of chunks in the master catalog.",
	})

	return &CatalogMetrics{
		ObjectCount: objectCount,
		ChunkCount: chunkCount,
	}
}
