package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joshjms/castletown/client"
)

func main() {
	// Create HTTP client
	c, err := client.NewHTTPClient("http://localhost:8000", nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Create a simple request using the builder
	req := client.NewRequest().
		AddFile("hello.txt", "Hello, World!").
		AddStep(func(p *client.ProcessBuilder) {
			p.WithImage("gcc:15-bookworm").
				WithCommand("/bin/cat", "hello.txt").
				WithFiles("hello.txt")
		}).
		Build()

	// Execute the request
	ctx := context.Background()
	resp, err := c.Execute(ctx, req)
	if err != nil {
		log.Fatalf("Failed to execute: %v", err)
	}

	// Print results
	fmt.Printf("Job ID: %s\n", resp.ID)
	for i, report := range resp.Reports {
		fmt.Printf("\nStep %d:\n", i+1)
		fmt.Printf("  Status: %s\n", report.Status)
		fmt.Printf("  Exit Code: %d\n", report.ExitCode)
		fmt.Printf("  Stdout: %s\n", report.Stdout)
		fmt.Printf("  Stderr: %s\n", report.Stderr)
		fmt.Printf("  CPU Time: %d ns\n", report.CPUTime)
		fmt.Printf("  Memory: %d bytes\n", report.Memory)
	}

	// Mark job as done (cleanup)
	if err := c.Done(ctx, resp.ID); err != nil {
		log.Printf("Warning: Failed to mark job as done: %v", err)
	}
}
