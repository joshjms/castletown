package client

// RequestBuilder provides a fluent API for building execution requests.
type RequestBuilder struct {
	req *ExecRequest
}

// NewRequest creates a new RequestBuilder.
func NewRequest() *RequestBuilder {
	return &RequestBuilder{
		req: &ExecRequest{
			Files: []File{},
			Steps: []Process{},
		},
	}
}

// WithID sets the job ID for the request.
func (b *RequestBuilder) WithID(id string) *RequestBuilder {
	b.req.ID = id
	return b
}

// AddFile adds a file to the request.
func (b *RequestBuilder) AddFile(name, content string) *RequestBuilder {
	b.req.Files = append(b.req.Files, File{
		Name:    name,
		Content: content,
	})
	return b
}

// AddStep adds a process/step to the request using a ProcessBuilder.
func (b *RequestBuilder) AddStep(fn func(*ProcessBuilder)) *RequestBuilder {
	pb := NewProcess()
	fn(pb)
	b.req.Steps = append(b.req.Steps, pb.Build())
	return b
}

// Build returns the constructed ExecRequest.
func (b *RequestBuilder) Build() *ExecRequest {
	return b.req
}

// ProcessBuilder provides a fluent API for building process specifications.
type ProcessBuilder struct {
	proc Process
}

// NewProcess creates a new ProcessBuilder.
func NewProcess() *ProcessBuilder {
	return &ProcessBuilder{
		proc: Process{
			Cmd:     []string{},
			Files:   []string{},
			Persist: []string{},
		},
	}
}

// WithImage sets the container image for the process.
func (p *ProcessBuilder) WithImage(image string) *ProcessBuilder {
	p.proc.Image = image
	return p
}

// WithCommand sets the command and arguments for the process.
func (p *ProcessBuilder) WithCommand(cmd ...string) *ProcessBuilder {
	p.proc.Cmd = cmd
	return p
}

// WithStdin sets the standard input for the process.
func (p *ProcessBuilder) WithStdin(stdin string) *ProcessBuilder {
	p.proc.Stdin = stdin
	return p
}

// WithMemoryLimit sets the memory limit in megabytes.
func (p *ProcessBuilder) WithMemoryLimit(mb int64) *ProcessBuilder {
	p.proc.MemoryLimitMB = mb
	return p
}

// WithTimeLimit sets the time limit in milliseconds.
func (p *ProcessBuilder) WithTimeLimit(ms uint64) *ProcessBuilder {
	p.proc.TimeLimitMs = ms
	return p
}

// WithProcLimit sets the maximum number of processes.
func (p *ProcessBuilder) WithProcLimit(limit int64) *ProcessBuilder {
	p.proc.ProcLimit = limit
	return p
}

// WithFiles specifies which files to make available in this step.
func (p *ProcessBuilder) WithFiles(files ...string) *ProcessBuilder {
	p.proc.Files = append(p.proc.Files, files...)
	return p
}

// WithPersist specifies which files to persist to the next step.
func (p *ProcessBuilder) WithPersist(files ...string) *ProcessBuilder {
	p.proc.Persist = append(p.proc.Persist, files...)
	return p
}

// Build returns the constructed Process.
func (p *ProcessBuilder) Build() Process {
	return p.proc
}

// Common helper functions for building requests

// SimpleExecRequest creates a simple single-step execution request.
// This is a convenience function for common use cases.
func SimpleExecRequest(image string, cmd []string, files map[string]string) *ExecRequest {
	req := NewRequest()

	// Add all files
	for name, content := range files {
		req.AddFile(name, content)
	}

	// Add single step with all files
	fileNames := make([]string, 0, len(files))
	for name := range files {
		fileNames = append(fileNames, name)
	}

	req.AddStep(func(p *ProcessBuilder) {
		p.WithImage(image).
			WithCommand(cmd...).
			WithFiles(fileNames...)
	})

	return req.Build()
}

// CompileAndRunRequest creates a two-step request for compile-then-run scenarios.
// This is useful for compiled languages like C++, Java, etc.
func CompileAndRunRequest(
	compileImage string,
	compileCmd []string,
	runImage string,
	runCmd []string,
	sourceFiles map[string]string,
	compiledOutputs []string,
	stdin string,
) *ExecRequest {
	req := NewRequest()

	// Add source files
	for name, content := range sourceFiles {
		req.AddFile(name, content)
	}

	// Get source file names
	sourceFileNames := make([]string, 0, len(sourceFiles))
	for name := range sourceFiles {
		sourceFileNames = append(sourceFileNames, name)
	}

	// Step 1: Compile
	req.AddStep(func(p *ProcessBuilder) {
		p.WithImage(compileImage).
			WithCommand(compileCmd...).
			WithFiles(sourceFileNames...).
			WithPersist(compiledOutputs...).
			WithMemoryLimit(512).
			WithTimeLimit(10000)
	})

	// Step 2: Run
	req.AddStep(func(p *ProcessBuilder) {
		p.WithImage(runImage).
			WithCommand(runCmd...).
			WithFiles(compiledOutputs...).
			WithStdin(stdin).
			WithMemoryLimit(256).
			WithTimeLimit(5000)
	})

	return req.Build()
}
