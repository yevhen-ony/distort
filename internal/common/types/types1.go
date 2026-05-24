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
	ObjectID ObjectID
	ChunkKey ChunkKey
}

type ChunkPlacement1 struct {
	Meta    ChunkMeta
	Slot    ObjectSlot
	Sources []NodeRef
}

type ChunkDesc1 struct {
	Placement ChunkPlacement1
}

type ObjectDesc1 struct {
	ID          ObjectID
	Size        int64
	ChunkCount  int
	Replication int
	Chunks      []ChunkPlacement1
}
