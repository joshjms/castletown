package sandbox_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joshjms/castletown/config"
	"github.com/joshjms/castletown/sandbox"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	sandbox.Init()
	config.UseDefaults()

	sandbox.NewManager()

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

func TestSandboxAdd(t *testing.T) {
	expectedStatus := sandbox.STATUS_OK
	expectedOutput := "15\n"

	tc := sandbox.Testcase{
		File:           "test_files/add.cpp",
		Stdin:          "6 9\n",
		ExpectedStatus: &expectedStatus,
		ExpectedOutput: &expectedOutput,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxTimeLimitExceededA(t *testing.T) {
	expectedStatus := sandbox.STATUS_TIME_LIMIT_EXCEEDED

	tc := sandbox.Testcase{
		File:           "test_files/tl1.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxTimeLimitExceededB(t *testing.T) {
	expectedStatus := sandbox.STATUS_TIME_LIMIT_EXCEEDED

	tc := sandbox.Testcase{
		File:           "test_files/printloop.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxMemoryLimitExceeded(t *testing.T) {
	expectedStatus := sandbox.STATUS_MEMORY_LIMIT_EXCEEDED

	tc := sandbox.Testcase{
		File:           "test_files/mem1.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    10000,
	}

	tc.Run(t)
}

func TestSandboxFork(t *testing.T) {
	expectedStatus := sandbox.STATUS_OK

	tc := sandbox.Testcase{
		File:           "test_files/fork.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	tc.Run(t)
}

func TestSandboxRusageConsistency(t *testing.T) {
	expectedStatus := sandbox.STATUS_OK

	tc := sandbox.Testcase{
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
