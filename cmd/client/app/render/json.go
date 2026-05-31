package render

import (
	"encoding/json"
	"errors"
	"io"

	"dos/cmd/client/app"
)

type JSONRender struct {
	json *json.Encoder
}

type Envelope struct {
	Operation string `json:"operation"`
	Error     string `json:"error,omitempty"`
	Result    any    `json:"result,omitempty"`
}

func NewJSONRender(out io.Writer, pretty bool) (*JSONRender, error) {
	if out == nil {
		return nil, errors.New("missing out")
	}
	enc := json.NewEncoder(out)
	if pretty {
		enc.SetIndent("", "  ")
	}

	render := &JSONRender{
		json: enc,
	}
	return render, nil
}

func (r *JSONRender) Error(op string, err error) error {
	return r.json.Encode(&Envelope{
		Operation: op,
		Error: err.Error(),
	})
}

func (r *JSONRender) Ping(res *app.PingResult) error {
	return r.json.Encode(&Envelope{
		Operation: "ping",
		Result: res,
	})
}

func (r *JSONRender) ListObjects(res *app.ListObjectsResult) error {
	return r.json.Encode(&Envelope{
		Operation: "list_objects",
		Result: res.Objects,
	})
}

func (r *JSONRender) ListChunks(res *app.ListChunksResult) error {
	return r.json.Encode(&Envelope{
		Operation: "list_chunks",
		Result: res.Chunks,
	})
}

func (r *JSONRender) ListNodes(res *app.ListNodesResult) error {
	return r.json.Encode(&Envelope{
		Operation: "list_nodes",
		Result: res.Nodes,
	})
}

func (r *JSONRender) DiscoverMaster(res *app.DiscoverMasterResult) error {
	return r.json.Encode(&Envelope{
		Operation: "leader",
		Result: res.MasterRef, 
	})
}
