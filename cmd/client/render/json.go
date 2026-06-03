package render

import (
	"encoding/json"

	"dos/cmd/client/app"
	"dos/internal/services/client/domain/progress"
)

type JSONRender struct {
	json *json.Encoder
}

type Envelope struct {
	Operation string `json:"operation"`
	Error     string `json:"error,omitempty"`
	Result    any    `json:"result,omitempty"`
}

func NewJSONRender() *JSONRender {
	return &JSONRender{}
}

func (r *JSONRender) Error(res *ErrorResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: res.Operation,
		Error:     res.Error.Error(),
	}, "", "  ")

}

func (r *JSONRender) Ping(res *app.PingResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "ping",
		Result:    res,
	}, "", "  ")
}

func (r *JSONRender) ListObjects(res *app.ListObjectsResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "list_objects",
		Result:    res.Objects,
	}, "", "  ")
}

func (r *JSONRender) ListChunks(res *app.ListChunksResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "list_chunks",
		Result:    res.Chunks,
	}, "", "  ")
}

func (r *JSONRender) ListNodes(res *app.ListNodesResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "list_nodes",
		Result:    res.Nodes,
	}, "", "  ")
}

func (r *JSONRender) DiscoverMaster(res *app.DiscoverMasterResult) ([]byte, error) {
	return json.Marshal(&Envelope{
		Operation: "leader",
		Result:    res.MasterRef,
	})
}

func (r *JSONRender) DescribeChunk(res *app.DescribeChunkResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "describe_chunk",
		Result:    res.Description.Placement,
	}, "", "  ")
}

func (r *JSONRender) DescribeObject(res *app.DescribeObjectResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "describe_object",
		Result:    res.Description,
	}, "", "  ")
}

func (r *JSONRender) DownloadChunk(res *app.DownloadChunkResult) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "download_chunk",
		Result: res,
	}, "", "  ")
}

func (r *JSONRender) AllocateChunk(res *app.AllocateChunkResult) ([]byte, error) {
  	return json.MarshalIndent(&Envelope{
  		Operation: "allocate_chunk",
  		Result:    res,
  	}, "", "  ")
}

func (r *JSONRender) PushChunk(res *app.PushChunkResult) ([]byte, error) {
  	return json.MarshalIndent(&Envelope{
  		Operation: "push_chunk",
  		Result:    res,
  	}, "", "  ")
}

func (r *JSONRender) CreateObject(res *app.CreateObjectResult) ([]byte, error) {
  	return json.MarshalIndent(&Envelope{
  		Operation: "create_object",
  		Result:    res,
  	}, "", "  ")
}

func (r *JSONRender) InspectNode(res *app.InspectNodeResult) ([]byte, error) {
  	return json.MarshalIndent(&Envelope{
  		Operation: "inspect_node",
  		Result:    res.Report,
  	}, "", "  ")
}

func (r *JSONRender) Progress(p *progress.ObjectProgress) ([]byte, error) {
	return json.MarshalIndent(&Envelope{
		Operation: "object_transfer_progress",
		Result: p, 
	}, "", "  ")
}
