package domain

import (
	m "dos/internal/services/master"
)

type MasterServiceConfig struct{
	ReplicationCount int
	ChunkAllocationMarginBytes int64
}

type MasterService struct {
	chunkRepo m.ChunkRepo
	objectRepo m.ObjectRepo
	nodeReg m.NodeRegistry

	index m.ChunkNodeIndex
	placementPolicy m.PlacementPolicy
	config *MasterServiceConfig
}

func NewMasterService(
	chunkRepo m.ChunkRepo,
	objectRepo m.ObjectRepo,
	nodeReg m.NodeRegistry,
	config *MasterServiceConfig,
) *MasterService {
	return &MasterService{
		chunkRepo: chunkRepo,
		objectRepo: objectRepo,
		nodeReg: nodeReg,
		placementPolicy: &RandomPlacementPolicy{},
		index: NewInMemChunkNodeIndex(),
		config: config,
	}
}

