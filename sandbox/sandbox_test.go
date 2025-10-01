package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshjms/castletown/config"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	Init()
	config.UseDefaults()

	files, err := os.ReadDir("test_files")
	require.NoError(nil, err, "failed to read test files directory: %v", err)

	for _, f := range files {
		fullPath := filepath.Join("test_files", f.Name())
		if err := os.Chown(fullPath, 0, 0); err != nil {
			panic(err)
		}
	}

	exitCode := m.Run()

	for _, f := range files {
		fullPath := filepath.Join("test_files", f.Name())
		if err := os.Chown(fullPath, 1000, 1000); err != nil {
			panic(err)
		}
	}

	os.Exit(exitCode)
}

func TestSandboxE2E(t *testing.T) {
	Init()

	m, err := NewManager()
	require.NoError(t, err, "failed to create manager: %v", err)

	id := uuid.NewString()
	rootFileDir := filepath.Join(config.StorageDir, id)
	defer os.RemoveAll(rootFileDir)

	proc0FileDir := filepath.Join(rootFileDir, "proc-0")
	os.MkdirAll(proc0FileDir, 0755)

	rootfsDir := "/tmp/_tmp_gcc_15-bookworm"

	cfg := &Config{
		RootfsImageDir: rootfsDir,
		BoxDir:         proc0FileDir,
		Args:           []string{"date", "+%s%3N"},
		Cwd:            "/box",
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		TimeLimitMs: 10000,
		Cgroup: &CgroupConfig{
			CpuQuota: 100000,
			Memory:   512 * 1024 * 1024,
		},
		Rlimit: &RlimitConfig{
			Core: &Rlimit{
				Hard: 0,
				Soft: 0,
			},
			Fsize: &Rlimit{
				Hard: 1024 * 1024 * 1024,
				Soft: 1024 * 1024 * 1024,
			},
			NoFile: &Rlimit{
				Hard: 64,
				Soft: 64,
			},
		},
	}

	sandbox, err := m.NewSandbox(fmt.Sprintf("%s-%d", id, 0), cfg)
	require.NoError(t, err, "failed to create sandbox: %v", err)
	defer m.DestroySandbox(sandbox.GetId())

	ctx := context.Background()

	execStartTime := time.Now()
	t.Logf("time now: %v", execStartTime.UnixMilli())
	report, err := sandbox.Run(ctx)

	ms, err := strconv.Atoi(strings.TrimSpace(report.Stdout))
	require.NoError(t, err, "failed to parse time in container: %v", err)
	t.Logf("time in container: %d\n", ms)

	timeDiff := int64(ms) - execStartTime.UnixMilli()
	t.Logf("time diff: %dms\n", timeDiff)

	require.NoError(t, err, "failed to run sandbox: %v", err)
	require.Equal(t, STATUS_OK, report.Status, "status not ok")
}

func TestSandboxAdd(t *testing.T) {
	Init()

	expectedStatus := STATUS_OK
	expectedOutput := "15\n"

	tc := Testcase{
		File:           "test_files/add.cpp",
		Stdin:          "6 9\n",
		ExpectedStatus: &expectedStatus,
		ExpectedOutput: &expectedOutput,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxTimeLimitExceededA(t *testing.T) {
	Init()

	expectedStatus := STATUS_TIME_LIMIT_EXCEEDED

	tc := Testcase{
		File:           "test_files/tl1.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxTimeLimitExceededB(t *testing.T) {
	Init()

	expectedStatus := STATUS_TIME_LIMIT_EXCEEDED

	tc := Testcase{
		File:           "test_files/printloop.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxMemoryLimitExceeded(t *testing.T) {
	Init()

	expectedStatus := STATUS_MEMORY_LIMIT_EXCEEDED

	tc := Testcase{
		File:           "test_files/mem1.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    10000,
	}

	tc.Run(t)
}

func TestSandboxFork(t *testing.T) {
	Init()

	expectedStatus := STATUS_OK

	tc := Testcase{
		File:           "test_files/fork.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxRusageConsistency(t *testing.T) {
	Init()

	expectedStatus := STATUS_OK

	tc := Testcase{
		File:           "test_files/random.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	var minCpuUsage, maxCpuUsage uint64

	for i := 0; i < 10; i++ {
		report := tc.Run(t)

		if i == 0 {
			minCpuUsage = report.CPUTime
			maxCpuUsage = report.CPUTime

			continue
		}

		minCpuUsage = min(minCpuUsage, report.CPUTime)
		maxCpuUsage = max(maxCpuUsage, report.CPUTime)
	}

	require.Less(t, maxCpuUsage-minCpuUsage, uint64(10000), "cpu usage inconsistent")
}
