package sandbox

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSandbox(t *testing.T) {
	Init()

	rootfsDir, err := createRootfs("gcc:11-bullseye")
	require.NoError(t, err, "failed to create rootfs directory: %v", err)
	defer os.RemoveAll(rootfsDir)

	config := &Config{
		Rootfs: rootfsDir,
		Args:   []string{"sh", "-c", "echo \"don't forget\""},
		Cwd:    "/",
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		UserNamespace: &UserNamespaceConfig{
			RootUID:     1000,
			UIDMapStart: 100000,
			UIDMapCount: 65534,
			RootGID:     1000,
			GIDMapStart: 100000,
			GIDMapCount: 65534,
		},
		Cgroup: &CgroupConfig{
			CpuQuota: 1000000,
			Memory:   512 * 1024 * 1024,
		},
	}

	sandbox := NewSandbox("sandbox-01", config)
	ctx := context.Background()

	err = sandbox.Init(ctx)
	require.NoError(t, err, "failed to init sandbox: %v", err)

	report, err := sandbox.Run(ctx)
	require.NoError(t, err, "failed to execute sandbox: %v", err)

	require.Equal(t, report.ExitCode, 0, "exit code != 0")
	require.Equal(t, report.Stdout, "don't forget\n", "stdout is different")

	t.Logf("Sandbox executed successfully: %+v", report)
}
