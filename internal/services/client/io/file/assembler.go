package file

import (
	"fmt"
	"os"
	"path/filepath"

	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

type ObjectAssembler struct {
	destPath string
}

func NewObjectAssembler(destPath string) (*ObjectAssembler, error) {
	destDir := filepath.Dir(destPath)
	err := os.MkdirAll(destDir, 0o755)
	if err != nil {
		return nil, err
	}

	asm := &ObjectAssembler{destPath: destPath}
	return asm, nil
}

func (a *ObjectAssembler) NewSink(chunks []t.ChunkPlacement1) (*ObjectSink, error) {
	layoutSpec := ObjectLayoutSpecFromChunkPlacments(chunks) 
	layout, err := NewObjectLayout(layoutSpec)	
	if err != nil {
		return nil, fmt.Errorf("create layout: %w", err)
	}

	writer, err := NewObjectWriter(a.destPath, layout.TotalBytes)
	if err != nil {
		return nil, fmt.Errorf("create writer: %w", err)
	}
	sink := &ObjectSink{
		writer: writer,
		layout: layout,
	}
	return sink, nil
}

type ObjectSink struct {
	writer *ObjectWriter
	layout *ObjectLayout
}

func (os *ObjectSink) WriteChunk(key t.ChunkKey, data []byte) error {
	region, err := os.layout.Region(key)
	if err != nil {
		return err 
	}
	return os.writer.WriteRegion(region, data)
}

func (os *ObjectSink) Close() error {
	return os.writer.Close()
}

func ObjectLayoutSpecFromChunkPlacments(chunks []t.ChunkPlacement1) *LayoutSpec {
	lcs := utils.Map(chunks, func(p t.ChunkPlacement1) LayoutChunk {
		return LayoutChunk {
			Key: p.Slot.ChunkKey, 
			Size: p.Meta.Digest.Size,
		}		
	})
	return &LayoutSpec{ chunks: lcs }
}


