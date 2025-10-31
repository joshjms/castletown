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

	// C++ source code
	cppSource := `#include <iostream>
#include <string>

int main() {
    std::string name;
    std::getline(std::cin, name);
    std::cout << "Hello, " << name << "!" << std::endl;
    return 0;
}`

	// Use the compile-and-run helper
	req := client.CompileAndRunRequest(
		"gcc:15-bookworm",              // Compile image
		[]string{"g++", "main.cpp", "-o", "main"}, // Compile command
		"gcc:15-bookworm",              // Run image
		[]string{"./main"},             // Run command
		map[string]string{              // Source files
			"main.cpp": cppSource,
		},
		[]string{"main"},               // Compiled outputs to persist
		"World",                        // Stdin for the program
	)

	// Execute the request
	ctx := context.Background()
	resp, err := c.Execute(ctx, req)
	if err != nil {
		log.Fatalf("Failed to execute: %v", err)
	}

	// Print results
	fmt.Printf("Job ID: %s\n", resp.ID)
	fmt.Println("\n--- Compilation Step ---")
	compileReport := resp.Reports[0]
	fmt.Printf("Status: %s\n", compileReport.Status)
	fmt.Printf("Exit Code: %d\n", compileReport.ExitCode)
	if compileReport.Stdout != "" {
		fmt.Printf("Stdout: %s\n", compileReport.Stdout)
	}
	if compileReport.Stderr != "" {
		fmt.Printf("Stderr: %s\n", compileReport.Stderr)
	}
	fmt.Printf("CPU Time: %d ns\n", compileReport.CPUTime)
	fmt.Printf("Memory: %d bytes\n", compileReport.Memory)

	fmt.Println("\n--- Execution Step ---")
	runReport := resp.Reports[1]
	fmt.Printf("Status: %s\n", runReport.Status)
	fmt.Printf("Exit Code: %d\n", runReport.ExitCode)
	fmt.Printf("Stdout: %s\n", runReport.Stdout)
	if runReport.Stderr != "" {
		fmt.Printf("Stderr: %s\n", runReport.Stderr)
	}
	fmt.Printf("CPU Time: %d ns\n", runReport.CPUTime)
	fmt.Printf("Memory: %d bytes\n", runReport.Memory)

	// Mark job as done (cleanup)
	if err := c.Done(ctx, resp.ID); err != nil {
		log.Printf("Warning: Failed to mark job as done: %v", err)
	}
}
