# Castletown Client Library

A comprehensive Go client library for interacting with the castletown sandboxed code execution service. This library provides both HTTP REST and gRPC clients with a unified interface.

## Features

- **Dual Protocol Support**: Connect via HTTP REST or gRPC
- **Unified Interface**: Same API regardless of protocol
- **Fluent Builder API**: Easy-to-use builder patterns for constructing requests
- **Type Safety**: Strongly-typed request/response structures
- **Context Support**: Full support for context-based cancellation and timeouts
- **Helper Functions**: Pre-built patterns for common use cases (compile-and-run, etc.)

## Installation

```bash
go get github.com/joshjms/castletown/client
```

## Quick Start

### HTTP Client

```go
package main

import (
    "context"
    "log"
    "github.com/joshjms/castletown/client"
)

func main() {
    // Create HTTP client
    c, err := client.NewHTTPClient("http://localhost:8000", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Build and execute request
    req := client.NewRequest().
        AddFile("hello.txt", "Hello, World!").
        AddStep(func(p *client.ProcessBuilder) {
            p.WithImage("gcc:15-bookworm").
              WithCommand("/bin/cat", "hello.txt").
              WithFiles("hello.txt")
        }).
        Build()

    resp, err := c.Execute(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Output: %s", resp.Reports[0].Stdout)
}
```

### gRPC Client

```go
// Create gRPC client
opts := &client.ClientOptions{
    GRPCOptions: &client.GRPCOptions{
        Insecure: true,
    },
}
c, err := client.NewGRPCClient("localhost:8001", opts)
if err != nil {
    log.Fatal(err)
}
defer c.Close()

// Use the same API as HTTP client
```

## Core Concepts

### Client Interface

The `Client` interface provides three main methods:

```go
type Client interface {
    // Execute submits a job and returns results
    Execute(ctx context.Context, req *ExecRequest) (*ExecResponse, error)

    // Done marks a job as complete (cleanup)
    Done(ctx context.Context, jobID string) error

    // Close releases client resources
    Close() error
}
```

### Execution Request

An execution request consists of:
- **Files**: Files to create in the sandbox
- **Steps**: Sequential processes to execute
- **ID**: Optional job identifier (auto-generated if not provided)

```go
type ExecRequest struct {
    ID    string
    Files []File
    Steps []Process
}
```

### Process Configuration

Each process step supports:
- **Image**: Container image name
- **Command**: Command and arguments
- **Stdin**: Standard input
- **Resource Limits**: Memory, time, and process limits
- **Files**: Which files to make available
- **Persist**: Which files to keep for next step

```go
type Process struct {
    Image         string
    Cmd           []string
    Stdin         string
    MemoryLimitMB int64
    TimeLimitMs   uint64
    ProcLimit     int64
    Files         []string
    Persist       []string
}
```

## Builder API

### RequestBuilder

The fluent builder API makes it easy to construct requests:

```go
req := client.NewRequest().
    WithID("my-job-id").
    AddFile("source.cpp", cppCode).
    AddStep(func(p *client.ProcessBuilder) {
        p.WithImage("gcc:15-bookworm").
          WithCommand("g++", "source.cpp", "-o", "program").
          WithFiles("source.cpp").
          WithPersist("program").
          WithMemoryLimit(512).
          WithTimeLimit(10000)
    }).
    AddStep(func(p *client.ProcessBuilder) {
        p.WithImage("gcc:15-bookworm").
          WithCommand("./program").
          WithFiles("program").
          WithStdin("test input").
          WithMemoryLimit(256).
          WithTimeLimit(5000)
    }).
    Build()
```

### ProcessBuilder

Configure individual process steps:

```go
p.WithImage("gcc:15-bookworm").           // Container image
  WithCommand("g++", "main.cpp").         // Command and args
  WithStdin("input data").                // Standard input
  WithMemoryLimit(512).                   // Memory limit (MB)
  WithTimeLimit(10000).                   // Time limit (ms)
  WithProcLimit(10).                      // Process limit
  WithFiles("main.cpp", "header.h").      // Available files
  WithPersist("main", "output.txt")       // Files to persist
```

## Helper Functions

### SimpleExecRequest

For simple single-step executions:

```go
req := client.SimpleExecRequest(
    "gcc:15-bookworm",                    // Image
    []string{"/bin/cat", "file.txt"},     // Command
    map[string]string{                    // Files
        "file.txt": "content",
    },
)
```

### CompileAndRunRequest

For compile-then-run workflows:

```go
req := client.CompileAndRunRequest(
    "gcc:15-bookworm",                           // Compile image
    []string{"g++", "main.cpp", "-o", "main"},   // Compile command
    "gcc:15-bookworm",                           // Run image
    []string{"./main"},                          // Run command
    map[string]string{"main.cpp": sourceCode},   // Source files
    []string{"main"},                            // Compiled outputs
    "test input",                                // Stdin for run
)
```

## Response Handling

### Execution Response

```go
type ExecResponse struct {
    ID      string   // Job ID
    Reports []Report // One per step
}
```

### Report Structure

Each report contains:

```go
type Report struct {
    Status   Status  // Execution status
    ExitCode int32   // Process exit code
    Signal   int32   // Termination signal (-1 if normal)
    Stdout   string  // Standard output
    Stderr   string  // Standard error
    CPUTime  uint64  // CPU time (nanoseconds)
    Memory   uint64  // Peak memory (bytes)
    WallTime int64   // Wall time (milliseconds)
    StartAt  int64   // Start timestamp (ns)
    FinishAt int64   // Finish timestamp (ns)
}
```

### Status Codes

```go
const (
    StatusOK                  // Successful execution
    StatusRuntimeError        // Program error
    StatusTimeLimitExceeded   // Time limit hit
    StatusMemoryLimitExceeded // Memory limit hit
    StatusOutputLimitExceeded // Output too large
    StatusTerminated          // Process terminated
    StatusUnknown             // Unknown error
    StatusSkipped             // Step skipped
)
```

## Configuration Options

### HTTP Client Options

```go
opts := &client.ClientOptions{
    Address: "http://localhost:8000",
    Timeout: 30 * time.Second,
}
c, err := client.NewHTTPClient("", opts)
```

### gRPC Client Options

```go
opts := &client.ClientOptions{
    Address: "localhost:8001",
    Timeout: 30 * time.Second,
    GRPCOptions: &client.GRPCOptions{
        Insecure:       true,        // Disable TLS
        MaxMessageSize: 4 * 1024 * 1024, // 4MB max message
    },
}
c, err := client.NewGRPCClient("", opts)
```

## Examples

See the `examples/` directory for complete working examples:

- **basic_http/**: Simple HTTP client example
- **basic_grpc/**: Simple gRPC client example
- **compile_and_run/**: Compile and run C++ code
- **advanced/**: Multi-step execution with resource limits

### Running Examples

```bash
# Start the castletown server first
cd /path/to/castletown
go run main.go server

# Run examples
cd client/examples/basic_http
go run main.go

cd ../compile_and_run
go run main.go

cd ../advanced
go run main.go
```

## Error Handling

Always check for errors and handle them appropriately:

```go
resp, err := c.Execute(ctx, req)
if err != nil {
    log.Fatalf("Execution failed: %v", err)
}

for i, report := range resp.Reports {
    if report.Status != client.StatusOK {
        log.Printf("Step %d failed: %s", i, report.Status)
        log.Printf("Stderr: %s", report.Stderr)
    }
}
```

## Context and Timeouts

Use contexts for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := c.Execute(ctx, req)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(10 * time.Second)
    cancel() // Cancel after 10 seconds
}()

resp, err := c.Execute(ctx, req)
```

## Resource Cleanup

Always clean up resources:

```go
// Close client when done
defer c.Close()

// Mark jobs as done for cleanup
resp, err := c.Execute(ctx, req)
if err == nil {
    defer c.Done(ctx, resp.ID)
}
```

## Best Practices

1. **Use context timeouts**: Always set appropriate timeouts
2. **Clean up jobs**: Call `Done()` after execution to free server resources
3. **Close clients**: Always `defer c.Close()` after creating clients
4. **Check status codes**: Verify `report.Status` for each step
5. **Handle errors gracefully**: Check stderr for error details
6. **Set resource limits**: Use `WithMemoryLimit()` and `WithTimeLimit()` to prevent runaway processes
7. **Persist selectively**: Only persist files needed for subsequent steps
8. **Use builders**: Leverage the fluent builder API for cleaner code

## Troubleshooting

### Connection Issues

- Verify server is running: `curl http://localhost:8000/exec`
- Check firewall settings
- For gRPC, ensure port 8001 is accessible

### Execution Failures

- Check `report.Stderr` for error messages
- Verify container image exists on server
- Ensure files are properly included with `WithFiles()`
- Check resource limits are sufficient

### Image Not Found

Images must be available on the server. See castletown docs for image preparation:

```bash
# On the server
skopeo copy docker://gcc:15-bookworm oci:images/gcc:15-bookworm
umoci unpack --image images/gcc:15-bookworm rootfs/gcc:15-bookworm
```

## API Reference

Full API documentation is available via godoc:

```bash
godoc -http=:6060
# Visit http://localhost:6060/pkg/github.com/joshjms/castletown/client/
```

## License

Same as castletown main project.

## Contributing

Contributions are welcome! Please submit issues and pull requests to the main castletown repository.
