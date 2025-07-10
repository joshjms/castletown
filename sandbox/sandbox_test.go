package sandbox

import (
	"context"
	"os"
	"testing"

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
		Cgroup: &CgroupConfig{
			CpuQuota: 1000000,
			Memory:   512 * 1024 * 1024,
		},
	}

	sandbox := NewSandbox("sandbox-01", config)
	ctx := context.Background()

	report, err := sandbox.Run(ctx)
	require.NoError(t, err, "failed to execute sandbox: %v", err)

	require.Equal(t, report.ExitCode, 0, "exit code != 0")
	require.Equal(t, report.Stdout, "don't forget\n", "stdout is different")

	t.Logf("Sandbox executed successfully: %+v", report)
}

func TestSandboxCompileExec(t *testing.T) {
	Init()

	rootfsDir := "/tmp/_tmp_gcc_15-bookworm"

	compileConfig := &Config{
		RootfsImageDir: rootfsDir,
		Args:           []string{"g++", "main.cpp", "-o", "main"},
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
		Cgroup: &CgroupConfig{
			CpuQuota: 1000000,
			Memory:   512 * 1024 * 1024,
		},
		Copy: []File{
			{
				Src: "test_files/hello.cpp",
				Dst: "/main.cpp",
			},
		},
		Save: []File{
			{
				Src: "/main",
				Dst: "./main",
			},
		},
	}

	compileSandbox := NewSandbox("sandbox-01", compileConfig)
	ctx := context.Background()

	compileReport, err := compileSandbox.Run(ctx)
	require.NoError(t, err, "failed to execute sandbox: %v", err)
	defer os.Remove("main")
	require.Equal(t, compileReport.ExitCode, 0, "exit code != 0")

	t.Logf("Compile executed successfully: %+v", compileReport)

	execConfig := &Config{
		RootfsImageDir: rootfsDir,
		Args:           []string{"./main"},
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
		Cgroup: &CgroupConfig{
			CpuQuota:  1000000,
			Memory:    512 * 1024 * 1024,
			PidsLimit: 1,
		},
		Copy: []File{
			{
				Src: "main",
				Dst: "main",
			},
		},
	}

	execSandbox := NewSandbox("sandbox-02", execConfig)
	ctx = context.Background()

	execReport, err := execSandbox.Run(ctx)
	require.NoError(t, err, "failed to execute sandbox: %v", err)

	require.Equal(t, execReport.ExitCode, 0, "exit code != 0")
	require.Equal(t, execReport.Stdout, "don't forget\n", "stdout is different")

	t.Logf("Exec executed successfully: %+v", execReport)
}
