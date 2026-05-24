package api

import (
	pb "dos/gen/proto/master/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateCreateObjectRequest(req *pb.CreateObjectRequest) error {
	if req.GetObjectId() == "" {
		return status.Error(codes.InvalidArgument, "missing object id")
	}
	return nil
}

func validateAllocateChunkRequest(req *pb.AllocateChunkRequest) error {
	slot := req.GetObjectSlot()
	if slot.GetObjectId() == "" {
		return status.Error(codes.InvalidArgument, "missing object id")
	}
	if slot.GetChunkKey() == "" {
		return status.Error(codes.InvalidArgument, "missing chunk key")
	}
	if req.GetChunkSize() <= 0 {
		return status.Error(codes.InvalidArgument, "invalid chunk size")
	}
	return nil
}

func validateGetObjectAccessRequest(req *pb.GetObjectAccessRequest) error {
	if req.GetObjectId() == "" {
		return status.Error(codes.InvalidArgument, "missing object id")
	}
	return nil
}

func validateRegisterStorageNodeRequest(req *pb.RegisterStorageNodeRequest) error {
	if req.GetAddr() == "" {
		return status.Error(codes.InvalidArgument, "missing addr")
	}
	return nil
}

func validateHeartbeatRequest(req *pb.HeartbeatRequest) error {
	if req.GetNodeId() == "" {
		return status.Error(codes.InvalidArgument, "missing node id")
	}
	if req.GetStats() == nil {
		return status.Error(codes.InvalidArgument, "missing stats")
	}
	return nil
}

func validateReportStorageRequest(req *pb.ReportStorageRequest) error {
	if req.GetNodeId() == "" {
		return status.Error(codes.InvalidArgument, "missing node id")
	}
	if len(req.GetReports()) == 0 {
		return status.Error(codes.InvalidArgument, "missing reports")
	}
	return nil
}

