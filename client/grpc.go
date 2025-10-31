package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/joshjms/castletown/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpcClient implements the Client interface using gRPC.
type grpcClient struct {
	conn       *grpc.ClientConn
	execClient pb.ExecServiceClient
	doneClient pb.DoneServiceClient
	timeout    time.Duration
}

// newGRPCClient creates a new gRPC client.
func newGRPCClient(address string, opts *ClientOptions) (*grpcClient, error) {
	// Build gRPC dial options
	dialOpts := []grpc.DialOption{}

	if opts.GRPCOptions.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if opts.GRPCOptions.MaxMessageSize > 0 {
		dialOpts = append(dialOpts,
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(opts.GRPCOptions.MaxMessageSize),
				grpc.MaxCallSendMsgSize(opts.GRPCOptions.MaxMessageSize),
			),
		)
	}

	// Connect to server
	conn, err := grpc.NewClient(address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &grpcClient{
		conn:       conn,
		execClient: pb.NewExecServiceClient(conn),
		doneClient: pb.NewDoneServiceClient(conn),
		timeout:    opts.Timeout,
	}, nil
}

// Execute submits a job for execution via gRPC.
func (c *grpcClient) Execute(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	// Set timeout if not already set in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	// Convert to protobuf format
	pbReq := &pb.ExecRequest{
		Id:    req.ID,
		Files: toProtoFiles(req.Files),
		Procs: toProtoProcesses(req.Steps),
	}

	// Call gRPC method
	pbResp, err := c.execClient.Execute(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC Execute failed: %w", err)
	}

	// Convert response
	return &ExecResponse{
		ID:      pbResp.Id,
		Reports: fromProtoReports(pbResp.Reports),
	}, nil
}

// Done notifies the server that a job is complete via gRPC.
func (c *grpcClient) Done(ctx context.Context, jobID string) error {
	// Set timeout if not already set in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	// Create request
	pbReq := &pb.DoneRequest{
		Id: jobID,
	}

	// Call gRPC method
	_, err := c.doneClient.Done(ctx, pbReq)
	if err != nil {
		return fmt.Errorf("gRPC Done failed: %w", err)
	}

	return nil
}

// Close closes the gRPC connection.
func (c *grpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
