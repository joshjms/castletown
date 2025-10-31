package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// httpClient implements the Client interface using HTTP REST API.
type httpClient struct {
	address string
	timeout time.Duration
	client  *http.Client
}

// httpExecRequest is the HTTP JSON request format for /exec endpoint.
type httpExecRequest struct {
	ID    string        `json:"id,omitempty"`
	Files []httpFile    `json:"files"`
	Steps []httpProcess `json:"steps"`
}

// httpFile is the HTTP JSON format for a file.
type httpFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// httpProcess is the HTTP JSON format for a process.
type httpProcess struct {
	Image         string   `json:"image"`
	Cmd           []string `json:"cmd"`
	Stdin         string   `json:"stdin,omitempty"`
	MemoryLimitMB int64    `json:"memoryLimitMB,omitempty"`
	TimeLimitMs   uint64   `json:"timeLimitMs,omitempty"`
	ProcLimit     int64    `json:"procLimit,omitempty"`
	Files         []string `json:"files,omitempty"`
	Persist       []string `json:"persist,omitempty"`
}

// httpExecResponse is the HTTP JSON response format for /exec endpoint.
type httpExecResponse struct {
	ID      string       `json:"id"`
	Reports []httpReport `json:"reports"`
}

// httpReport is the HTTP JSON format for a report.
type httpReport struct {
	Status   string `json:"Status"`
	ExitCode int32  `json:"ExitCode"`
	Signal   int32  `json:"Signal"`
	Stdout   string `json:"Stdout"`
	Stderr   string `json:"Stderr"`
	CPUTime  uint64 `json:"CPUTime"`
	Memory   uint64 `json:"Memory"`
	WallTime int64  `json:"WallTime"`
	StartAt  int64  `json:"StartAt"`
	FinishAt int64  `json:"FinishAt"`
}

// httpDoneRequest is the HTTP JSON request format for /done endpoint.
type httpDoneRequest struct {
	ID string `json:"id"`
}

// Execute submits a job for execution via HTTP REST API.
func (c *httpClient) Execute(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	// Convert to HTTP format
	httpReq := httpExecRequest{
		ID:    req.ID,
		Files: make([]httpFile, len(req.Files)),
		Steps: make([]httpProcess, len(req.Steps)),
	}

	for i, f := range req.Files {
		httpReq.Files[i] = httpFile(f)
	}

	for i, p := range req.Steps {
		httpReq.Steps[i] = httpProcess(p)
	}

	// Marshal to JSON
	body, err := json.Marshal(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpRequest, err := http.NewRequestWithContext(ctx, "POST", c.address+"/exec", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	// Send request
	if c.client == nil {
		c.client = &http.Client{
			Timeout: c.timeout,
		}
	}

	resp, err := c.client.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var httpResp httpExecResponse
	if err := json.NewDecoder(resp.Body).Decode(&httpResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to client format
	response := &ExecResponse{
		ID:      httpResp.ID,
		Reports: make([]Report, len(httpResp.Reports)),
	}

	for i, r := range httpResp.Reports {
		response.Reports[i] = Report{
			Status:   parseStatus(r.Status),
			ExitCode: r.ExitCode,
			Signal:   r.Signal,
			Stdout:   r.Stdout,
			Stderr:   r.Stderr,
			CPUTime:  r.CPUTime,
			Memory:   r.Memory,
			WallTime: r.WallTime,
			StartAt:  r.StartAt,
			FinishAt: r.FinishAt,
		}
	}

	return response, nil
}

// Done notifies the server that a job is complete via HTTP REST API.
func (c *httpClient) Done(ctx context.Context, jobID string) error {
	// Create request
	httpReq := httpDoneRequest{
		ID: jobID,
	}

	body, err := json.Marshal(httpReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpRequest, err := http.NewRequestWithContext(ctx, "POST", c.address+"/done", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	// Send request
	if c.client == nil {
		c.client = &http.Client{
			Timeout: c.timeout,
		}
	}

	resp, err := c.client.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Close closes the HTTP client (no-op for HTTP).
func (c *httpClient) Close() error {
	return nil
}

// parseStatus converts a string status to Status enum.
func parseStatus(s string) Status {
	switch s {
	case "OK":
		return StatusOK
	case "RUNTIME_ERROR":
		return StatusRuntimeError
	case "TIME_LIMIT_EXCEEDED":
		return StatusTimeLimitExceeded
	case "MEMORY_LIMIT_EXCEEDED":
		return StatusMemoryLimitExceeded
	case "OUTPUT_LIMIT_EXCEEDED":
		return StatusOutputLimitExceeded
	case "TERMINATED":
		return StatusTerminated
	case "SKIPPED":
		return StatusSkipped
	case "UNKNOWN":
		return StatusUnknown
	default:
		return StatusUnspecified
	}
}
