package sandbox

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

type Status string

const (
	STATUS_OK                    Status = "OK"
	STATUS_RUNTIME_ERROR         Status = "RUNTIME_ERROR"
	STATUS_TIME_LIMIT_EXCEEDED   Status = "TIME_LIMIT_EXCEEDED"
	STATUS_MEMORY_LIMIT_EXCEEDED Status = "MEMORY_LIMIT_EXCEEDED"
	STATUS_OUTPUT_LIMIT_EXCEEDED Status = "OUTPUT_LIMIT_EXCEEDED"
	STATUS_TERMINATED            Status = "TERMINATED"
	STATUS_UNKNOWN               Status = "UNKNOWN"
)

type Report struct {
	Status   Status
	ExitCode int
	Signal   syscall.Signal
	Stdout   string
	Stderr   string
	CPUTime  uint64
	Memory   uint64
	WallTime int64
}

func (s *Sandbox) makeReport(stdoutBuf, stderrBuf io.Reader, state *os.ProcessState, timeLimitExceeded bool) (*Report, error) {
	stdout, err := io.ReadAll(stdoutBuf)
	if err != nil {
		return nil, fmt.Errorf("error reading stdout: %w", err)
	}

	stderr, err := io.ReadAll(stderrBuf)
	if err != nil {
		return nil, fmt.Errorf("error reading stderr: %w", err)
	}

	cgManager, err := loadCgroup(s.id)
	if err != nil {
		return nil, fmt.Errorf("error loading cgroup: %w", err)
	}

	stats, err := cgManager.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting cgroup stats: %w", err)
	}

	var status Status

	switch {
	case timeLimitExceeded:
		status = STATUS_TIME_LIMIT_EXCEEDED
	case stats.GetMemory().GetMaxUsage() > uint64(s.config.Cgroup.Memory):
		status = STATUS_MEMORY_LIMIT_EXCEEDED
	default:
		status = STATUS_OK
	}

	return &Report{
		Status:   status,
		ExitCode: state.ExitCode(),
		Signal:   state.Sys().(syscall.WaitStatus).Signal(),
		Stdout:   string(stdout),
		Stderr:   string(stderr),
		CPUTime:  stats.GetCPU().GetUserUsec(),
		Memory:   stats.GetMemory().GetMaxUsage(),
	}, nil
}
