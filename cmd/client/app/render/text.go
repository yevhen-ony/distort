package render

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"dos/cmd/client/app"
)

type TextRender struct {
	out io.Writer
}

func NewTextRender(out io.Writer) (*TextRender, error) {
	if out == nil {
		return nil, errors.New("missing out")
	}
	render := &TextRender{
		out: out,
	}
	return render, nil
}

func (r *TextRender) Error(opName string, opErr error) error {
	_, err := fmt.Fprintf(r.out,
		"%s:\n\t * error: %s\n",
		strings.ToUpper(opName),
		opErr.Error(),
	)
	return err
}

func (r *TextRender) Ping(res *app.PingResult) error {
	b := &strings.Builder{}

	fmt.Fprintln(b, "PING:")
	fmt.Fprintf(b, "\t * address  : %s\n", res.Address)
	fmt.Fprintf(b, "\t * status   : %s\n", res.Status)
	fmt.Fprintf(b, "\t * component: %s\n", res.Component)

	_, err := fmt.Fprint(r.out, b.String())
	return err
}

func (r *TextRender) ListObjects(res *app.ListObjectsResult) error {
	b := &strings.Builder{}

	fmt.Fprintf(b, "%-20s %11s %11s\n", "OBJECT_ID", "CHUNK_COUNT", "REPLICATION")
	for _, info := range res.Objects {
		fmt.Fprintf(b, "%-20s %11d %11d\n",
			info.ID,
			info.ChunkCount,
			info.Replication,
		)
	}

	_, err :=fmt.Fprint(r.out, b.String())
	return err
}

func (r *TextRender) ListChunks(res *app.ListChunksResult) error {
	b := &strings.Builder{}

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

	_, err := fmt.Fprint(r.out, b.String())
	return err
}

func (r *TextRender) ListNodes(res *app.ListNodesResult) error {
	b := &strings.Builder{}

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
	_, err := fmt.Fprint(r.out, b.String())
	return err
}

func (r *TextRender) DiscoverMaster(res *app.DiscoverMasterResult) error {

	b := &strings.Builder{}
	fmt.Fprintf(b, "%-20s %-20s\n", "MASTER_ID", "MASTER_ADDR")
	fmt.Fprintf(b, "%-20s %-20s\n", res.MasterRef.ID, res.MasterRef.Addr)

	_, err := fmt.Print(b.String())
	return err
	
}

func ToMBStr(bytes int64) string {
	mb := float64(bytes) / float64(1024 * 1024)
	return fmt.Sprintf("%.1fMB", mb)
}
