package sandbox

import (
	"context"
	"os"
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

func (tc *Testcase) Run(t *testing.T) {
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
		TimeLimitMs: tc.TimeLimitMs,
		Cgroup: &CgroupConfig{
			CpuQuota: 100000,
			Memory:   512 * 1024 * 1024,
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

	compileSandbox := NewSandbox(uuid.New().String(), compileConfig)
	ctx := context.Background()

	compileReport, err := compileSandbox.Run(ctx)
	defer os.Remove("main")

	require.NoError(t, err, "failed to compile code")
	require.Equal(t, STATUS_OK, compileReport.Status, "status not ok")

	t.Logf("%v\n", compileReport)

	execConfig := &Config{
		RootfsImageDir: rootfsDir,
		Args:           []string{"./main"},
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
		TimeLimitMs: 1000,
		Cgroup: &CgroupConfig{
			CpuQuota:  100000,
			Memory:    256 * 1024 * 1024,
			PidsLimit: 1,
		},
		Copy: []File{
			{
				Src: "main",
				Dst: "/box/main",
			},
		},
	}

	execSandbox := NewSandbox(uuid.New().String(), execConfig)
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
}
