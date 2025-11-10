package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type Testcase struct {
	File  string
	Stdin string

	ExpectedStatus *Status
	ExpectedOutput *string

	TimeLimitMs int64

	Concurrency int
}

func (tc *Testcase) Run(t *testing.T) []Report {
	m := GetManager()
	require.NotNil(t, m, "failed to get manager")

	id := uuid.NewString()
	rootFileDir := filepath.Join(m.cfg.StorageDir, id)
	defer os.RemoveAll(rootFileDir)

	compileFileDir := filepath.Join(rootFileDir, "proc-0")
	os.MkdirAll(compileFileDir, 0755)

	rootfsDir := filepath.Join(m.cfg.ImagesDir, "gcc-15-bookworm")

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

	compileId := fmt.Sprintf("%s-%d", id, 0)
	err := m.NewSandbox(compileId, compileConfig)
	defer require.NoError(t, err, "failed to create compile sandbox: %v", err)
	defer m.DestroySandbox(compileId)

	ctx := context.Background()
	compileStartTime := time.Now()
	compileReport, err := m.RunSandbox(ctx, compileId)
	require.NoError(t, err, "failed to compile code")
	compileElapsed := time.Since(compileStartTime)
	t.Logf("Compile took %v", compileElapsed)

	require.Equal(t, STATUS_OK, compileReport.Status, "compile status not ok")

	if tc.Concurrency < 1 {
		tc.Concurrency = 1
	}

	wg := sync.WaitGroup{}

	finishTimes := make([]time.Time, tc.Concurrency)
	reports := make([]Report, tc.Concurrency)

	for i := 1; i <= tc.Concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			execId := fmt.Sprintf("%s-%d", id, i)

			execFileDir := filepath.Join(rootFileDir, fmt.Sprintf("proc-%d", i))
			os.MkdirAll(execFileDir, 0755)

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

			err = m.NewSandbox(execId, execConfig)
			defer m.DestroySandbox(execId)
			require.NoError(t, err, "failed to create exec sandbox: %v", err)

			ctx = context.Background()
			execStartTime := time.Now()
			t.Logf("Starting execution %d at %v", i, execStartTime)
			execReport, err := m.RunSandbox(ctx, execId)
			execFinishTime := time.Now()
			execElapsed := time.Since(execStartTime)
			t.Logf("Finished execution %d at %v", i, execFinishTime)
			finishTimes[i-1] = execFinishTime
			t.Logf("Execution %d took %v", i, execElapsed)
			require.NoError(t, err, "failed to execute code")

			if tc.ExpectedStatus != nil {
				require.Equal(t, *tc.ExpectedStatus, execReport.Status, "status != expectedStatus")
			}

			if tc.ExpectedOutput != nil {
				require.Equal(t, *tc.ExpectedOutput, execReport.Stdout, "output != expectedOutput")
			}

			reports[i-1] = execReport
		}(i)
	}

	wg.Wait()

	return reports
}
