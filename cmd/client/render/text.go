package render

import (
	"bytes"
	"fmt"
	"strings"

	"dos/cmd/client/app"
	"dos/internal/services/client/domain/progress"
)

type TextRender struct{}

func NewTextRender() *TextRender {
	return &TextRender{}
}

func (r *TextRender) Error(res *ErrorResult) ([]byte, error) {
	s := fmt.Sprintf(
		"%s:\n\t * error: %s\n",
		strings.ToUpper(res.Operation),
		res.Error.Error(),
	)
	return []byte(s), nil
}

func (r *TextRender) Ping(res *app.PingResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintln(b, "PING:")
	fmt.Fprintf(b, "\t * address  : %s\n", res.Address)
	fmt.Fprintf(b, "\t * status   : %s\n", res.Status)
	fmt.Fprintf(b, "\t * component: %s\n", res.Component)

	return b.Bytes(), nil
}

func (r *TextRender) ListObjects(res *app.ListObjectsResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintf(b, "%-20s %11s %11s\n", "OBJECT_ID", "CHUNK_COUNT", "REPLICATION")
	for _, info := range res.Objects {
		fmt.Fprintf(b, "%-20s %11d %11d\n",
			info.ID,
			info.ChunkCount,
			info.Replication,
		)
	}
	return b.Bytes(), nil
}

func (r *TextRender) ListChunks(res *app.ListChunksResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintf(b,
		"%-18s %-8s %-8s %-20s\n",
		"CHUNK_ID", "SIZE", "REPLICAS", "OBJECT_ID",
	)
	for _, info := range res.Chunks {
		fmt.Fprintf(b,
			"%-18s %8s %8d %-20s\n",
			info.ID,
			ToMBStr(info.Size),
			info.ReplicaCount,
			info.ObjectID,
		)
	}
	return b.Bytes(), nil
}

func (r *TextRender) ListNodes(res *app.ListNodesResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintf(b,
		"%-18s %-18s %-6s %-8s\n",
		"NODE_ID", "ADDR", "CHUNKS", "SIZE",
	)

	for _, info := range res.Nodes {
		fmt.Fprintf(b,
			"%-18s %-18s %6d %8s\n",
			info.ID,
			info.Addr,
			info.ChunkCount,
			ToMBStr(info.UsedBytes),
		)
	}
	return b.Bytes(), nil
}

func (r *TextRender) DiscoverMaster(res *app.DiscoverMasterResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintf(b, "%-20s %-20s\n", "MASTER_ID", "MASTER_ADDR")
	fmt.Fprintf(b, "%-20s %-20s\n", res.MasterRef.ID, res.MasterRef.Addr)

	return b.Bytes(), nil
}

func (r *TextRender) DescribeChunk(res *app.DescribeChunkResult) ([]byte, error) {
	placement := res.Description.Placement

	b := &bytes.Buffer{}

	meta := placement.Meta
	fmt.Fprintln(b, "CHUNK META:")
	fmt.Fprintf(b, "\t * chunk_id: %s\n", meta.ID)
	fmt.Fprintf(b, "\t * checksum: %s\n", meta.Digest.Checksum)
	fmt.Fprintf(b, "\t * size    : %s\n", ToMBStr(meta.Digest.Size))

	slot := placement.Slot
	fmt.Fprintln(b, "OBJECT SLOT:")
	fmt.Fprintf(b, "\t * object_id: %s\n", slot.ObjectID)
	fmt.Fprintf(b, "\t * chunk_key: %s\n", slot.ChunkKey)

	sources := placement.Sources
	fmt.Fprintf(b, "SOURCES (STORAGE NODES) %d:\n", len(sources))
	for _, ref := range sources {
		fmt.Fprintf(b, "\t * node_id: %s | node_addr: %s\n", ref.ID, ref.Addr)
	}
	return b.Bytes(), nil
}

func (r *TextRender) DescribeObject(res *app.DescribeObjectResult) ([]byte, error) {
	desc := res.Description

	b := &bytes.Buffer{}

	fmt.Fprintln(b, "OBJECT:")
	fmt.Fprintf(b, "\t * object_id  : %s\n", desc.ID)
	fmt.Fprintf(b, "\t * total_size : %s\n", ToMBStr(desc.Size))
	fmt.Fprintf(b, "\t * chunks     : %d\n", len(desc.Chunks))
	fmt.Fprintf(b, "\t * replication: %d\n", desc.Replication)

	fmt.Fprintln(b, "CHUNKS:")
	fmt.Fprintf(b, "%-10s %-18s %11s %8s\n",
		"KEY",
		"CHUNK_ID",
		"SIZE",
		"REPLICAS",
	)

	for _, chunk := range desc.Chunks {

		fmt.Fprintf(b, "%-10s %-18s %10s %8d\n",
			chunk.Slot.ChunkKey,
			chunk.Meta.ID,
			ToMBStr(chunk.Meta.Digest.Size),
			len(chunk.Sources),
		)
	}
	return b.Bytes(), nil
}

func (r *TextRender) DownloadChunk(res *app.DownloadChunkResult) ([]byte, error) {

	b := &bytes.Buffer{}

	fmt.Fprintln(b, "GET CHUNK:")
	fmt.Fprintf(b, "\t * chunk_id : %s\n", res.Meta.ID)
	fmt.Fprintf(b, "\t * size     : %s\n", ToMBStr(res.Meta.Digest.Size))
	fmt.Fprintf(b, "\t * checksum : %s\n", res.Meta.Digest.Checksum)
	fmt.Fprintf(b, "\t * node_id  : %s\n", res.Source.ID)
	fmt.Fprintf(b, "\t * node_addr: %s\n", res.Source.Addr)
	fmt.Fprintf(b, "\t * path     : %s\n", res.Path)

	return b.Bytes(), nil
}

func (r *TextRender) AllocateChunk(res *app.AllocateChunkResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintln(b, "ALLOCATE CHUNK:")
	fmt.Fprintf(b, "\t * object_id: %s\n", res.ObjectID)
	fmt.Fprintf(b, "\t * chunk_key: %s\n", res.ChunkKey)
	fmt.Fprintf(b, "\t * chunk_id : %s\n", res.ChunkID)

	fmt.Fprintf(b, "TARGETS %d:\n", len(res.Targets))
	fmt.Fprintf(b, "%18s %18s\n", "NODE_ID", "ADDR")
	for _, target := range res.Targets {
		fmt.Fprintf(b, "%18s %18s\n", target.ID, target.Addr)
	}

	return b.Bytes(), nil
}

func (r *TextRender) PushChunk(res *app.PushChunkResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintln(b, "PUSH CHUNK:")
	fmt.Fprintf(b, "\t * chunk_id : %s\n", res.Meta.ID)
	fmt.Fprintf(b, "\t * size     : %s\n", ToMBStr(res.Meta.Digest.Size))
	fmt.Fprintf(b, "\t * checksum : %s\n", res.Meta.Digest.Checksum)
	fmt.Fprintf(b, "\t * node_id  : %s\n", res.Target.ID)
	fmt.Fprintf(b, "\t * node_addr: %s\n", res.Target.Addr)
	fmt.Fprintf(b, "\t * file     : %s\n", res.File)

	return b.Bytes(), nil
}

func (r *TextRender) CreateObject(res *app.CreateObjectResult) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintln(b, "CREATE OBJECT:")
	fmt.Fprintf(b, "\t * object_id: %s\n", res.ObjectID)

	return b.Bytes(), nil
}

func (r *TextRender) InspectNode(res *app.InspectNodeResult) ([]byte, error) {
  	report := res.Report
  	b := &bytes.Buffer{}

  	fmt.Fprintln(b, "INSPECT NODE:")
  	fmt.Fprintf(b, "\t * addr      : %s\n", report.Addr)
  	fmt.Fprintf(b, "\t * chunks    : %d\n", report.Stats.ChunkCount)
  	fmt.Fprintf(b, "\t * used_bytes: %s\n", ToMBStr(report.Stats.UsedBytes))
  	fmt.Fprintf(b, "\t * free_bytes: %s\n", ToMBStr(report.Stats.FreeBytes))
	fmt.Fprintf(b, "\t * heartbeat : status = %s\n", res.Report.Heartbeat.Status)

  	fmt.Fprintf(b, "%-18s %10s %-10s\n", "CHUNK_ID", "SIZE", "STATE")
  	for _, chunk := range report.Chunks {
  		fmt.Fprintf(b, "%-18s %10s %-10s\n",
  			chunk.Meta.ID,
  			ToMBStr(chunk.Meta.Digest.Size),
  			chunk.State,
  		)
  	}

  	return b.Bytes(), nil
}

func (r *TextRender) TriggerReport(res *app.TriggerReportResult) ([]byte, error) {
  	report := res.Report
  	b := &bytes.Buffer{}

  	fmt.Fprintln(b, "TRIGGER REPORT:")
  	if len(report.Scheduled) > 0 {
  		fmt.Fprintf(b, "SCHEDULED %d:\n", len(report.Scheduled))
  		for _, chunkID := range report.Scheduled {
  			fmt.Fprintf(b, "%-18s\n", chunkID)
  		}
  	}

  	if len(report.Failed) > 0 {
  		fmt.Fprintf(b, "FAILED %d:\n", len(report.Failed))
  		for _, chunkID := range report.Failed {
  			fmt.Fprintf(b, "%-18s\n", chunkID)
  		}
  	}

  	return b.Bytes(), nil
}

func (r *TextRender) HeartbeatControl(res *app.HeartbeatControlResult) ([]byte, error) {
	report := res.Report

	b := &bytes.Buffer{}
	fmt.Fprintln(b, "HEARTBEAT:")
	fmt.Fprintf(b, "\t * addr  : %s\n", report.Addr)
	fmt.Fprintf(b, "\t * status: %s\n", report.Heartbeat.Status)
	
	return b.Bytes(), nil
}


func (r *TextRender) Progress(op *progress.ObjectProgress) ([]byte, error) {
	b := &bytes.Buffer{}

	fmt.Fprintf(b, "OBJECT: %s\n", op.ObjectID)
	fmt.Fprintf(b, "STATUS: %s\n", op.GetStatusStr())

	fmt.Fprintf(b,
		"%-10s %-20s %-10s %-10s %-6s\n",
		"KEY", "ID", "SIZE", "SENT", "STATUS",
	)

	for _, key := range op.ChunksOrder {
		ch, ok := op.Chunks[key]
		if !ok {
			continue
		}

		sizeMB := ToMBStr(ch.Meta.Digest.Size)
		sentMB := ToMBStr(ch.SentBytes)
		fmt.Fprintf(
			b,
			"%-10s %-20s %10s %10s %10s\n",
			key, ch.Meta.ID, sizeMB, sentMB, ch.GetStatusStr(),
		)
	}
	return b.Bytes(), nil
}



func ToMBStr(bytes int64) string {
	mb := float64(bytes) / float64(1024*1024)
	return fmt.Sprintf("%.1fMB", mb)
}
