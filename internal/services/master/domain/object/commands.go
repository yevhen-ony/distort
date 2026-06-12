package object

import (
	"errors"

	t "dos/internal/common/types"
)

var (
	ErrInvalidObjectCommand = errors.New("invalid object command")
)

type ObjectCommand struct {
	CreateObject   *CreateObjectCommand   `json:"create_object,omitempty"`
	DeleteObject   *DeleteObjectCommand   `json:"delete_object,omitempty"`
	AddChunk       *AddChunkCommand       `json:"add_chunk,omitempty"`
	DeleteChunk    *DeleteChunkCommand    `json:"delete_chunk,omitempty"`
	SetReplication *SetReplicationCommand `json:"set_replication,omitempty"`
}

type CreateObjectCommand struct {
	ObjectID    t.ObjectID `json:"object_id"`
	Replication int        `json:"replication"`
}

func (cmd *ObjectCommand) Validate() error {
	count := 0
	for _, present := range []bool{
		cmd.CreateObject != nil,
		cmd.DeleteObject != nil,
		cmd.AddChunk != nil,
		cmd.DeleteChunk != nil,
		cmd.SetReplication != nil,
	} {
		if present {
			count++
		}
	}
	if count != 1 {
		return ErrInvalidObjectCommand
	}
	return nil
}

func (cmd *CreateObjectCommand) ToCommand() ObjectCommand {
	return ObjectCommand{CreateObject: cmd}
}

type DeleteObjectCommand struct {
	ObjectID t.ObjectID `json:"object_id"`
}

func (cmd *DeleteObjectCommand) ToCommand() ObjectCommand {
	return ObjectCommand{DeleteObject: cmd}
}

type SetReplicationCommand struct {
	ObjectID    t.ObjectID `json:"object_id"`
	Replication int        `json:"replication"`
}

func (cmd *SetReplicationCommand) ToCommand() ObjectCommand {
	return ObjectCommand{SetReplication: cmd}
}

type AddChunkCommand struct {
	ObjectID t.ObjectID `json:"object_id"`
	ChunkKey t.ChunkKey `json:"chunk_key"`
	ChunkID  t.ChunkID  `json:"chunk_id"`
}

func (cmd *AddChunkCommand) ToCommand() ObjectCommand {
	return ObjectCommand{AddChunk: cmd}
}

type DeleteChunkCommand struct {
	ObjectID t.ObjectID `json:"object_id"`
	ChunkKey t.ChunkKey `json:"chunk_key"`
}

func (cmd *DeleteChunkCommand) ToCommand() ObjectCommand {
	return ObjectCommand{DeleteChunk: cmd}
}
