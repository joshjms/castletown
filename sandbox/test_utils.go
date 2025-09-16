package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type Testcase struct {
	File  string
	Stdin string

	ExpectedStatus *Status
	ExpectedOutput *string

	TimeLimitMs int64
}

func (tc *Testcase) Run(t *testing.T) *Report {
	m, err := NewManager()
	require.NoError(t, err, "failed to create manager: %v", err)

	rootfsDir := "/tmp/_tmp_gcc_15-bookworm"

	compileConfig := &Config{
		RootfsImageDir: rootfsDir,
		Args:           []string{"g++", "main.cpp", "-o", "main"},
		Cwd:            "/box",
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		ContainerUID: 65534,
		ContainerGID: 65534,
		UserNamespace: &UserNamespaceConfig{
			RootUID:     uint32(os.Getuid()),
			UIDMapStart: 100000,
			UIDMapCount: 65534,
			RootGID:     uint32(os.Getgid()),
			GIDMapStart: 100000,
			GIDMapCount: 65534,
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
				Hard: 1 * 1024 * 1024,
				Soft: 1 * 1024 * 1024,
			},
			NoFile: &Rlimit{
				Hard: 64,
				Soft: 64,
			},
		},
		Copy: []File{
			{
				Src: tc.File,
				Dst: "/box/main.cpp",
			},
		},
		Save: []File{
			{
				Src: "/box/main",
				Dst: "./main",
			},
		},
	}

	id := uuid.New().String()
	compileSandbox, err := m.NewSandbox(id, compileConfig, filepath.Join("/tmp", id))
	require.NoError(t, err, "failed to create compile sandbox: %v", err)
	ctx := context.Background()

	compileReport, err := compileSandbox.Run(ctx)
	defer os.Remove("main")

	require.NoError(t, err, "failed to compile code")
	require.Equal(t, STATUS_OK, compileReport.Status, "status not ok")

	t.Logf("%v\n", compileReport)

	execConfig := &Config{
		RootfsImageDir: rootfsDir,
		Args:           []string{"./main"},
		Stdin:          tc.Stdin,
		Cwd:            "/box",
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		ContainerUID: 65534,
		ContainerGID: 65534,
		UserNamespace: &UserNamespaceConfig{
			RootUID:     uint32(os.Getuid()),
			UIDMapStart: 100000,
			UIDMapCount: 65534,
			RootGID:     uint32(os.Getgid()),
			GIDMapStart: 100000,
			GIDMapCount: 65534,
		},
		TimeLimitMs: tc.TimeLimitMs,
		Cgroup: &CgroupConfig{
			CpuQuota:   100000,
			Memory:     256 * 1024 * 1024,
			PidsLimit:  1,
			CpusetCpus: "0",
			CpusetMems: "0",
		},
		Copy: []File{
			{
				Src: "main",
				Dst: "/box/main",
			},
		},
	}

	id = uuid.NewString()
	execSandbox, err := m.NewSandbox(id, execConfig, filepath.Join("/tmp", id))
	require.NoError(t, err, "failed to create exec sandbox: %v", err)
	ctx = context.Background()

	execReport, err := execSandbox.Run(ctx)
	t.Logf("%v\n", execReport)
	require.NoError(t, err, "failed to execute code")

	if tc.ExpectedStatus != nil {
		require.Equal(t, *tc.ExpectedStatus, execReport.Status, "status != expectedStatus")
	}

	if tc.ExpectedOutput != nil {
		require.Equal(t, *tc.ExpectedOutput, execReport.Stdout, "output != expectedOutput")
	}

	return execReport
}
