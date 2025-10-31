package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joshjms/castletown/client"
)

func main() {
	// Create HTTP client with custom timeout
	opts := &client.ClientOptions{
		Timeout: 60 * time.Second,
	}
	c, err := client.NewHTTPClient("http://localhost:8000", opts)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Python source code
	pythonSource := `import sys

def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

n = int(input())
result = factorial(n)
print(f"Factorial of {n} is {result}")
`

	// Create a request with resource limits
	req := client.NewRequest().
		WithID("advanced-example-job"). // Custom job ID
		AddFile("factorial.py", pythonSource).
		AddFile("input.txt", "10").
		AddStep(func(p *client.ProcessBuilder) {
			p.WithImage("gcc:15-bookworm").
				WithCommand("/bin/cat", "input.txt").
				WithFiles("input.txt").
				WithPersist("input.txt").          // Persist to next step
				WithMemoryLimit(128).              // 128 MB limit
				WithTimeLimit(2000).               // 2 second limit
				WithProcLimit(10)                  // Max 10 processes
		}).
		AddStep(func(p *client.ProcessBuilder) {
			p.WithImage("gcc:15-bookworm").
				WithCommand("python3", "factorial.py").
				WithFiles("factorial.py", "input.txt").
				WithStdin("10"). // Alternative: use stdin directly
				WithMemoryLimit(256).
				WithTimeLimit(5000).
				WithProcLimit(20)
		}).
		Build()

	// Execute with context
	ctx := context.Background()
	resp, err := c.Execute(ctx, req)
	if err != nil {
		log.Fatalf("Failed to execute: %v", err)
	}

	// Print detailed results
	fmt.Printf("Job ID: %s\n", resp.ID)
	fmt.Printf("Total steps: %d\n", len(resp.Reports))

	for i, report := range resp.Reports {
		fmt.Printf("\n=== Step %d ===\n", i+1)
		fmt.Printf("Status: %s\n", report.Status)
		fmt.Printf("Exit Code: %d\n", report.ExitCode)

		if report.Signal != -1 {
			fmt.Printf("Signal: %d\n", report.Signal)
		}

		// Check for errors
		switch report.Status {
		case client.StatusOK:
			fmt.Println("✓ Execution successful")
		case client.StatusRuntimeError:
			fmt.Println("✗ Runtime error")
		case client.StatusTimeLimitExceeded:
			fmt.Println("✗ Time limit exceeded")
		case client.StatusMemoryLimitExceeded:
			fmt.Println("✗ Memory limit exceeded")
		default:
			fmt.Printf("✗ Status: %s\n", report.Status)
		}

		// Output
		if report.Stdout != "" {
			fmt.Printf("\nStdout:\n%s\n", report.Stdout)
		}
		if report.Stderr != "" {
			fmt.Printf("\nStderr:\n%s\n", report.Stderr)
		}

		// Resource usage
		fmt.Printf("\nResource Usage:\n")
		fmt.Printf("  CPU Time: %.2f ms\n", float64(report.CPUTime)/1_000_000)
		fmt.Printf("  Memory: %.2f MB\n", float64(report.Memory)/1_048_576)
		fmt.Printf("  Wall Time: %d ms\n", report.WallTime)

		// Timing
		duration := time.Duration(report.FinishAt - report.StartAt)
		fmt.Printf("  Total Duration: %v\n", duration)
	}

	// Mark job as done
	fmt.Println("\nCleaning up...")
	if err := c.Done(ctx, resp.ID); err != nil {
		log.Printf("Warning: Failed to mark job as done: %v", err)
	} else {
		fmt.Println("Job marked as done successfully")
	}
}
