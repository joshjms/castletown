package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	pb "github.com/joshjms/castletown/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	httpURL      = "http://localhost:8000"
	grpcAddr     = "localhost:8001"
	defaultImage = "gcc:15-bookworm"
)

type HTTPExecRequest struct {
	ID    string        `json:"id"`
	Files []HTTPFile    `json:"files"`
	Steps []HTTPProcess `json:"steps"`
}

type HTTPFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type HTTPProcess struct {
	Image         string   `json:"image"`
	Cmd           []string `json:"cmd"`
	Stdin         string   `json:"stdin"`
	MemoryLimitMB int64    `json:"memoryLimitMB"`
	TimeLimitMs   uint64   `json:"timeLimitMs"`
	ProcLimit     int64    `json:"procLimit"`
	Files         []string `json:"files"`
	Persist       []string `json:"persist"`
}

type HTTPExecResponse struct {
	ID      string       `json:"id"`
	Reports []HTTPReport `json:"reports"`
}

type HTTPReport struct {
	Status   string `json:"Status"`
	ExitCode int    `json:"ExitCode"`
	Signal   int    `json:"Signal"`
	Stdout   string `json:"Stdout"`
	Stderr   string `json:"Stderr"`
	CPUTime  uint64 `json:"CPUTime"`
	Memory   uint64 `json:"Memory"`
	WallTime int64  `json:"WallTime"`
}

type HTTPDoneRequest struct {
	ID string `json:"id"`
}

func TestHTTPExec(t *testing.T) {
	// Create a simple exec request
	req := HTTPExecRequest{
		ID: "test-http-exec",
		Files: []HTTPFile{
			{
				Name:    "test.txt",
				Content: "Hello from HTTP test",
			},
		},
		Steps: []HTTPProcess{
			{
				Image: defaultImage,
				Cmd:   []string{"/bin/cat", "test.txt"},
				Files: []string{"test.txt"},
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(httpURL+"/exec", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var execResp HTTPExecResponse
	if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if execResp.ID != "test-http-exec" {
		t.Errorf("Expected ID 'test-http-exec', got '%s'", execResp.ID)
	}

	if len(execResp.Reports) == 0 {
		t.Fatal("Expected at least one report")
	}

	report := execResp.Reports[0]
	if report.Status != "OK" {
		t.Errorf("Expected status OK, got %s", report.Status)
	}

	if report.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", report.ExitCode)
	}

	t.Logf("HTTP Exec test passed. Stdout: %s", report.Stdout)
}

func TestHTTPDone(t *testing.T) {
	req := HTTPDoneRequest{
		ID: "test-http-exec",
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(httpURL+"/done", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	t.Log("HTTP Done test passed")
}

func TestGRPCExec(t *testing.T) {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewExecServiceClient(conn)

	req := &pb.ExecRequest{
		Id: "test-grpc-exec",
		Files: []*pb.File{
			{
				Name:    "test.txt",
				Content: "Hello from gRPC test",
			},
		},
		Procs: []*pb.Process{
			{
				Image: defaultImage,
				Cmd:   []string{"/bin/cat", "test.txt"},
				Files: []string{"test.txt"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute gRPC request: %v", err)
	}

	if resp.Id != "test-grpc-exec" {
		t.Errorf("Expected ID 'test-grpc-exec', got '%s'", resp.Id)
	}

	if len(resp.Reports) == 0 {
		t.Fatal("Expected at least one report")
	}

	report := resp.Reports[0]
	if report.Status != pb.Status_STATUS_OK {
		t.Errorf("Expected status OK, got %v", report.Status)
	}

	if report.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", report.ExitCode)
	}

	t.Logf("gRPC Exec test passed. Stdout: %s", report.Stdout)
}

func TestGRPCDone(t *testing.T) {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewDoneServiceClient(conn)

	req := &pb.DoneRequest{
		Id: "test-grpc-exec",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Done(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute gRPC done request: %v", err)
	}

	if resp == nil {
		t.Error("Expected non-nil response")
	}

	t.Log("gRPC Done test passed")
}

func TestHTTPAndGRPCIntegration(t *testing.T) {
	httpReq := HTTPExecRequest{
		ID: "test-integration",
		Files: []HTTPFile{
			{
				Name:    "input.txt",
				Content: "Hello, world!",
			},
		},
		Steps: []HTTPProcess{
			{
				Image: defaultImage,
				Cmd:   []string{"/bin/echo", "Hello, world!"},
				Files: []string{},
			},
		},
	}

	reqBody, _ := json.Marshal(httpReq)
	httpResp, err := http.Post(httpURL+"/exec", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to send HTTP request: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		t.Fatalf("HTTP request failed with status %d: %s", httpResp.StatusCode, string(body))
	}

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	doneClient := pb.NewDoneServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = doneClient.Done(ctx, &pb.DoneRequest{Id: "test-integration"})
	if err != nil {
		t.Fatalf("Failed to mark job done via gRPC: %v", err)
	}

	t.Log("Integration test passed: HTTP exec + gRPC done")
}
