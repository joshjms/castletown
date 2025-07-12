package sandbox

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSandboxRun(t *testing.T) {
	Init()

	rootfsDir := "/tmp/_tmp_gcc_15-bookworm"

	config := &Config{
		RootfsImageDir: rootfsDir,
		Args:           []string{"sh", "-c", "echo \"don't forget\""},
		Cwd:            "/",
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		UserNamespace: &UserNamespaceConfig{
			RootUID:     uint32(os.Getuid()),
			UIDMapStart: 100000,
			UIDMapCount: 65534,
			RootGID:     uint32(os.Getgid()),
			GIDMapStart: 100000,
			GIDMapCount: 65534,
		},
		TimeLimitMs: 1000,
		Cgroup: &CgroupConfig{
			CpuQuota: 1000000,
			Memory:   512 * 1024 * 1024,
		},
	}

	sandbox := NewSandbox(uuid.New().String(), config)
	ctx := context.Background()

	report, err := sandbox.Run(ctx)
	require.NoError(t, err, "failed to execute sandbox: %v", err)

	require.Equal(t, report.ExitCode, 0, "exit code != 0")
	require.Equal(t, report.Stdout, "don't forget\n", "stdout is different")

	t.Logf("Sandbox executed successfully: %+v", report)
}

func TestSandboxCompileExec(t *testing.T) {
	Init()

	expectedStatus := STATUS_OK
	expectedOutput := "don't forget\n"

	tc := Testcase{
		File:           "test_files/hello.cpp",
		ExpectedStatus: &expectedStatus,
		ExpectedOutput: &expectedOutput,
		TimeLimitMs:    1000,
	}

	require.NoError(t, tc.Run())
}

func TestSandboxTimeLimitExceeded(t *testing.T) {
	Init()

	expectedStatus := STATUS_TIME_LIMIT_EXCEEDED

	tc := Testcase{
		File:           "test_files/tle.cpp",
		ExpectedStatus: &expectedStatus,
		TimeLimitMs:    1000,
	}

	require.NoError(t, tc.Run())
}
