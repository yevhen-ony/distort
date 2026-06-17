package render

import (
	"dos/cmd/client/app"
	"dos/internal/services/client/domain/progress"
	"errors"
)

type ErrorResult struct {
	Operation string
	Error     error
}

func NewErrorResult(op string, err error) *ErrorResult {
	return &ErrorResult{
		Operation: op,
		Error:     err,
	}
}

type Render interface {
	Error(*ErrorResult) ([]byte, error)

	Ping(*app.PingResult) ([]byte, error)
	DiscoverMaster(*app.DiscoverMasterResult) ([]byte, error)

	ListNodes(*app.ListNodesResult) ([]byte, error)
	InspectNode(*app.InspectNodeResult) ([]byte, error)
	TriggerReport(*app.TriggerReportResult) ([]byte, error)
	HeartbeatControl(*app.HeartbeatControlResult) ([]byte, error)

	DescribeObject(*app.DescribeObjectResult) ([]byte, error)
	ListObjects(*app.ListObjectsResult) ([]byte, error)
	CreateObject(*app.CreateObjectResult) ([]byte, error)

	DownloadChunk(*app.DownloadChunkResult) ([]byte, error)
	AllocateChunk(*app.AllocateChunkResult) ([]byte, error)
	PushChunk(*app.PushChunkResult) ([]byte, error)
	ListChunks(*app.ListChunksResult) ([]byte, error)
	DescribeChunk(*app.DescribeChunkResult) ([]byte, error)

	Progress(*progress.ObjectProgress) ([]byte, error)
}

func NewRender(format string) (Render, error) {
	switch format {
	case "json":
		return NewJSONRender(), nil
	case "text":
		return NewTextRender(), nil
	default:
		return nil, errors.New("unknown output format")
	}
}
