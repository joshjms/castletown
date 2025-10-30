package done

import (
	"context"

	"github.com/joshjms/castletown/job"
	pb "github.com/joshjms/castletown/proto"
)

type DoneServer struct {
	pb.UnimplementedDoneServiceServer
}

func NewDoneServer() *DoneServer {
	return &DoneServer{}
}

func (s *DoneServer) Done(ctx context.Context, req *pb.DoneRequest) (*pb.DoneResponse, error) {
	jp := job.GetJobPool()
	jp.RemoveJob(req.Id)

	return &pb.DoneResponse{}, nil
}
