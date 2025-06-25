package castletown

import (
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup2/stats"
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

func MakeReport(stdoutBuf, stderrBuf io.Reader, state *os.ProcessState, stats *stats.Metrics) (*Report, error) {
	stdout, err := io.ReadAll(stdoutBuf)
	if err != nil {
		return nil, fmt.Errorf("error reading stdout: %w", err)
	}

	stderr, err := io.ReadAll(stderrBuf)
	if err != nil {
		return nil, fmt.Errorf("error reading stderr: %w", err)
	}

	r := &Report{
		Status:   STATUS_OK,
		ExitCode: state.ExitCode(),
		Signal:   state.Sys().(syscall.WaitStatus).Signal(),
		Stdout:   string(stdout),
		Stderr:   string(stderr),
		CPUTime:  stats.GetCPU().UserUsec,
		Memory:   stats.GetMemory().Usage,
	}

	return r, nil
}
