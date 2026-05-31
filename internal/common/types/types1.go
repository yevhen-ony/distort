package types

type HealthStatus string

const (
	HealthUnknown     HealthStatus = "unknown"
	HealthOK          HealthStatus = "ok"
	HealthDegraded    HealthStatus = "degraded"
	HealthUnavailable HealthStatus = "unavailable"
)

type ChunkAllocation1 struct {
	ID      ChunkID
	Slot    ObjectSlot
	Targets []NodeRef
}

type ObjectSlot struct {
	ObjectID ObjectID `json:"object_id"`
	ChunkKey ChunkKey `json:"chunk_key"`
}

type ChunkPlacement1 struct {
	Meta    ChunkMeta  `json:"chunk_meta"`
	Slot    ObjectSlot `json:"object_slog"`
	Sources []NodeRef  `json:"sources"`
}

type ChunkDesc1 struct {
	Placement ChunkPlacement1 `json:"chunk_placement"`
}

type ObjectDesc1 struct {
	ID          ObjectID          `json:"object_id"`
	Size        int64             `json:"total_size"`
	Replication int               `json:"replication"`
	Chunks      []ChunkPlacement1 `json:"chunks"`
}
