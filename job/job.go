package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/joshjms/castletown/sandbox"
)

type Job struct {
	ID    string    `json:"id"`
	Files []File    `json:"files"`
	Procs []Process `json:"steps"`

	step int

	mu sync.Mutex
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Process struct {
	Image         string   `json:"image"`
	Cmd           []string `json:"cmd"`
	Stdin         string   `json:"stdin"`
	MemoryLimitMB int64    `json:"memoryLimitMB"`
	TimeLimitMs   uint64   `json:"timeLimitMs"`
	ProcLimit     int64    `json:"procLimit"`
	Files         []string `json:"files"`
	Persist       []string `json:"persist"`
}

func (j *Job) Prepare() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if len(j.Procs) == 0 {
		return fmt.Errorf("no processes specified")
	}

	if err := verifyImages(j.Procs); err != nil {
		return fmt.Errorf("invalid images: %w", err)
	}

	if err := prepareFileDirs(j.ID, j.Procs); err != nil {
		return fmt.Errorf("error preparing file directories: %w", err)
	}

	return nil
}

func (j *Job) ExecuteAll(ctx context.Context) ([]sandbox.Report, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	var reports []sandbox.Report

	for j.step < len(j.Procs) {
		report, err := j.execute(ctx)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (j *Job) execute(ctx context.Context) (sandbox.Report, error) {
	proc := j.Procs[j.step]
	fileDeps, err := getFileDependencies(j.ID, j.Procs, j.Files, j.step)
	if err != nil {
		return sandbox.Report{}, fmt.Errorf("error getting file dependencies: %w", err)
	}

	cfg := sandbox.GetDefaultConfig()
	cfg.Args = proc.Cmd
	cfg.RootfsImageDir = getImageDir(proc.Image)
	cfg.BoxDir = getProcFileDir(j.ID, j.step)
	cfg.Files = fileDeps

	if proc.TimeLimitMs > 0 {
		cfg.TimeLimitMs = int64(proc.TimeLimitMs)
	}
	if proc.MemoryLimitMB > 0 {
		cfg.Cgroup.Memory = int64(proc.MemoryLimitMB) * 1024 * 1024
	}
	if proc.ProcLimit > 0 {
		cfg.Cgroup.PidsLimit = proc.ProcLimit
	}
	cfg.Stdin = proc.Stdin

	containerId := fmt.Sprintf("%s-%d", j.ID, j.step)
	s, err := sandbox.GetManager().NewSandbox(containerId, cfg)
	defer sandbox.GetManager().DestroySandbox(containerId)

	if err != nil {
		return sandbox.Report{}, fmt.Errorf("cannot create sandbox for process %d: %v", j.step, err)
	}

	report, err := s.Run(ctx)
	if err != nil {
		return sandbox.Report{}, fmt.Errorf("error running process %d: %v", j.step, err)
	}

	j.next()

	return report, nil
}

func (j *Job) next() bool {
	if j.step < len(j.Procs) {
		j.step++
		return true
	}
	return false
}
