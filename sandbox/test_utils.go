package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshjms/castletown/config"
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
	m := GetManager()
	require.NotNil(t, m, "failed to get manager")

	id := uuid.NewString()
	rootFileDir := filepath.Join(config.StorageDir, id)
	defer os.RemoveAll(rootFileDir)

	compileFileDir := filepath.Join(rootFileDir, "proc-0")
	execFileDir := filepath.Join(rootFileDir, "proc-1")

	os.MkdirAll(compileFileDir, 0755)
	os.MkdirAll(execFileDir, 0755)

	rootfsDir := "/tmp/_tmp_gcc_15-bookworm"

	compileConfig := &Config{
		RootfsImageDir: rootfsDir,
		BoxDir:         compileFileDir,
		Args:           []string{"g++", "-o", "main", "main.cpp"},
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
		Files: []File{
			{
				Src: tc.File,
				Dst: filepath.Join(compileFileDir, "main.cpp"),
			},
		},
	}

	compileSandbox, err := m.NewSandbox(fmt.Sprintf("%s-%d", id, 0), compileConfig)
	defer require.NoError(t, err, "failed to create compile sandbox: %v", err)
	defer m.DestroySandbox(compileSandbox.GetId())

	ctx := context.Background()
	compileStartTime := time.Now()
	compileReport, err := compileSandbox.Run(ctx)
	require.NoError(t, err, "failed to compile code")
	compileElapsed := time.Since(compileStartTime)
	t.Logf("Compile took %v", compileElapsed)

	require.Equal(t, STATUS_OK, compileReport.Status, "compile status not ok")

	execConfig := &Config{
		RootfsImageDir: rootfsDir,
		BoxDir:         execFileDir,
		Args:           []string{"./main"},
		Stdin:          tc.Stdin,
		Cwd:            "/box",
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		TimeLimitMs: tc.TimeLimitMs,
		Cgroup: &CgroupConfig{
			CpuQuota:   100000,
			Memory:     256 * 1024 * 1024,
			PidsLimit:  1,
			CpusetCpus: "0",
			CpusetMems: "0",
		},
		Files: []File{
			{
				Src: filepath.Join(compileFileDir, "main"),
				Dst: filepath.Join(execFileDir, "main"),
			},
		},
	}

	execSandbox, err := m.NewSandbox(fmt.Sprintf("%s-%d", id, 1), execConfig)
	defer m.DestroySandbox(execSandbox.GetId())
	require.NoError(t, err, "failed to create exec sandbox: %v", err)

	ctx = context.Background()
	execStartTime := time.Now()
	execReport, err := execSandbox.Run(ctx)
	execElapsed := time.Since(execStartTime)
	t.Logf("Execution took %v", execElapsed)
	require.NoError(t, err, "failed to execute code")

	if tc.ExpectedStatus != nil {
		require.Equal(t, *tc.ExpectedStatus, execReport.Status, "status != expectedStatus")
	}

	if tc.ExpectedOutput != nil {
		require.Equal(t, *tc.ExpectedOutput, execReport.Stdout, "output != expectedOutput")
	}

	return execReport
}
