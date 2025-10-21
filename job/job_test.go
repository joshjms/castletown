package job_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joshjms/castletown/config"
	"github.com/joshjms/castletown/job"
	"github.com/joshjms/castletown/sandbox"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	sandbox.Init()
	config.UseDefaults()

	job.NewJobPool()
	sandbox.NewManager(config.MaxConcurrency)

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestJob(t *testing.T) {
	var err error

	jobId := uuid.NewString()
	j := &job.Job{
		ID: jobId,
		Procs: []job.Process{
			{
				Image:   "gcc:15-bookworm",
				Cmd:     []string{"g++", "-o", "main", "main.cpp"},
				Files:   []string{"main.cpp"},
				Persist: []string{"main"},
			},
			{
				Image: "gcc:15-bookworm",
				Cmd:   []string{"./main"},
				Stdin: "5\n",
				Files: []string{"main"},
			},
		},
		Files: []job.File{
			{
				Name: "main.cpp",
				Content: `
#include <iostream>
int main() {
	int n;
	std::cin >> n;
	std::cout << n * n << std::endl;
	return 0;
}
`,
			},
		},
	}

	err = j.Prepare()
	require.NoError(t, err, "error preparing job: %v", err)

	reports, err := j.ExecuteAll(context.Background())
	require.NoError(t, err, "error executing job: %v", err)
	require.Len(t, reports, 2, "expected 2 reports, got %d", len(reports))
	require.Equal(t, sandbox.STATUS_OK, reports[0].Status, "expected first report status to be OK, got %v", reports[0].Status)
	require.Equal(t, sandbox.STATUS_OK, reports[1].Status, "expected second report status to be OK, got %v", reports[1].Status)
	require.Equal(t, "25\n", reports[1].Stdout, "expected second report output to be '25', got '%s'", reports[1].Stdout)
}

func TestJobAppend(t *testing.T) {
	var err error

	pool := job.GetJobPool()

	jobId := uuid.NewString()
	firstJob := &job.Job{
		ID: jobId,
		Procs: []job.Process{
			{
				Image:   "gcc:15-bookworm",
				Cmd:     []string{"g++", "-o", "main", "main.cpp"},
				Files:   []string{"main.cpp"},
				Persist: []string{"main"},
			},
			{
				Image: "gcc:15-bookworm",
				Cmd:   []string{"./main"},
				Stdin: "5\n",
				Files: []string{"main"},
			},
		},
		Files: []job.File{
			{
				Name: "main.cpp",
				Content: `
#include <iostream>
int main() {
	int n;
	std::cin >> n;
	std::cout << n * n << std::endl;
	return 0;
}
`,
			},
		},
	}

	pool.AddOrAppendJob(firstJob)
	err = firstJob.Prepare()
	require.NoError(t, err, "error preparing job: %v", err)

	reports, err := firstJob.ExecuteAll(context.Background())

	require.NoError(t, err, "error executing job: %v", err)
	require.Len(t, reports, 2, "expected 2 reports, got %d", len(reports))
	require.Equal(t, sandbox.STATUS_OK, reports[0].Status, "expected first report status to be OK, got %v", reports[0].Status)
	require.Equal(t, sandbox.STATUS_OK, reports[1].Status, "expected second report status to be OK, got %v", reports[1].Status)
	require.Equal(t, "25\n", reports[1].Stdout, "expected second report output to be '25', got '%s'", reports[1].Stdout)

	anotherJob := &job.Job{
		ID: jobId,
		Procs: []job.Process{
			{
				Image: "gcc:15-bookworm",
				Cmd:   []string{"./main"},
				Stdin: "10\n",
				Files: []string{"main"},
			},
		},
	}

	pool.AddOrAppendJob(anotherJob)

	pooledJob, exists := pool.Jobs[jobId]
	require.True(t, exists, "expected job to exist in pool")
	require.Len(t, pooledJob.Procs, 3, "expected 3 processes in pooled job, got %d", len(pooledJob.Procs))

	reports, err = pooledJob.ExecuteAll(context.Background())

	require.NoError(t, err, "error executing pooled job: %v", err)
	require.Len(t, reports, 1, "expected 1 report, got %d", len(reports))
	require.Equal(t, sandbox.STATUS_OK, reports[0].Status, "expected first report status to be OK, got %v", reports[0].Status)
	require.Equal(t, "100\n", reports[0].Stdout, "expected first report output to be '100', got '%s'", reports[0].Stdout)

	pool.RemoveJob(jobId)
	_, exists = pool.Jobs[jobId]
	require.False(t, exists, "expected job to be removed from pool")
}
