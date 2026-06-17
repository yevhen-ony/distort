package types

type HealthStatus string

const (
	HealthUnknown     HealthStatus = "unknown"
	HealthOK          HealthStatus = "ok"
	HealthDegraded    HealthStatus = "degraded"
	HealthUnavailable HealthStatus = "unavailable"
)

type ChunkAllocation struct {
	ID      ChunkID
	Slot    ObjectSlot
	Targets []NodeRef
}

type ObjectSlot struct {
	ObjectID ObjectID `json:"object_id"`
	ChunkKey ChunkKey `json:"chunk_key"`
}

type ChunkPlacement struct {
	Meta    ChunkMeta  `json:"chunk_meta"`
	Slot    ObjectSlot `json:"object_slog"`
	Sources []NodeRef  `json:"sources"`
}

type ChunkDesc struct {
	Placement ChunkPlacement `json:"chunk_placement"`
}

type ObjectDesc struct {
	ID          ObjectID         `json:"object_id"`
	Size        int64            `json:"total_size"`
	Replication int              `json:"replication"`
	Chunks      []ChunkPlacement `json:"chunks"`
}
