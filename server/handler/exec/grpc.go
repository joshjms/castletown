package exec

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshjms/castletown/job"
	pb "github.com/joshjms/castletown/proto"
	"github.com/joshjms/castletown/sandbox"
)

type ExecServer struct {
	pb.UnimplementedExecServiceServer
}

func NewExecServer() *ExecServer {
	return &ExecServer{}
}

func (s *ExecServer) Execute(ctx context.Context, req *pb.ExecRequest) (*pb.ExecResponse, error) {
	id := req.Id
	if id == "" {
		id = uuid.NewString()
	}

	files := make([]job.File, len(req.Files))
	for i, f := range req.Files {
		files[i] = job.File{
			Name:    f.Name,
			Content: f.Content,
		}
	}

	procs := make([]job.Process, len(req.Procs))
	for i, p := range req.Procs {
		procs[i] = job.Process{
			Image:         p.Image,
			Cmd:           p.Cmd,
			Stdin:         p.Stdin,
			MemoryLimitMB: p.MemoryLimitMb,
			TimeLimitMs:   p.TimeLimitMs,
			ProcLimit:     p.ProcLimit,
			Files:         p.Files,
			Persist:       p.Persist,
		}
	}

	apiReq := Request{
		ID:    id,
		Files: files,
		Procs: procs,
	}

	reports, err := handleRequest(ctx, apiReq)
	if err != nil {
		return nil, err
	}

	protoReports := make([]*pb.Report, len(reports))
	for i, r := range reports {
		protoReports[i] = convertToProtoReport(r)
	}

	return &pb.ExecResponse{
		Id:      id,
		Reports: protoReports,
	}, nil
}

func convertToProtoReport(r sandbox.Report) *pb.Report {
	return &pb.Report{
		Status:   convertToProtoStatus(r.Status),
		ExitCode: int32(r.ExitCode),
		Signal:   int32(r.Signal),
		Stdout:   r.Stdout,
		Stderr:   r.Stderr,
		CpuTime:  r.CPUTime,
		Memory:   r.Memory,
		WallTime: r.WallTime,
		StartAt:  r.StartAt.UnixNano(),
		FinishAt: r.FinishAt.UnixNano(),
	}
}

func convertToProtoStatus(status sandbox.Status) pb.Status {
	switch status {
	case sandbox.STATUS_OK:
		return pb.Status_STATUS_OK
	case sandbox.STATUS_RUNTIME_ERROR:
		return pb.Status_STATUS_RUNTIME_ERROR
	case sandbox.STATUS_TIME_LIMIT_EXCEEDED:
		return pb.Status_STATUS_TIME_LIMIT_EXCEEDED
	case sandbox.STATUS_MEMORY_LIMIT_EXCEEDED:
		return pb.Status_STATUS_MEMORY_LIMIT_EXCEEDED
	case sandbox.STATUS_OUTPUT_LIMIT_EXCEEDED:
		return pb.Status_STATUS_OUTPUT_LIMIT_EXCEEDED
	case sandbox.STATUS_TERMINATED:
		return pb.Status_STATUS_TERMINATED
	case sandbox.STATUS_SKIPPED:
		return pb.Status_STATUS_SKIPPED
	default:
		return pb.Status_STATUS_UNKNOWN
	}
}
