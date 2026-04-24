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
	if req.GetObjectId() == "" {
		return status.Error(codes.InvalidArgument, "missing object id")
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
