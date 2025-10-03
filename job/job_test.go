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

	sandbox.NewManager()

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestJob(t *testing.T) {
	var err error

	jobId := uuid.NewString()
	job := &job.Job{
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

	err = job.Prepare()
	require.NoError(t, err, "error preparing job: %v", err)

	reports, err := job.ExecuteAll(context.Background())
	require.NoError(t, err, "error executing job: %v", err)
	require.Len(t, reports, 2, "expected 2 reports, got %d", len(reports))
	require.Equal(t, sandbox.STATUS_OK, reports[0].Status, "expected first report status to be OK, got %v", reports[0].Status)
	require.Equal(t, sandbox.STATUS_OK, reports[1].Status, "expected second report status to be OK, got %v", reports[1].Status)
	require.Equal(t, "25\n", reports[1].Stdout, "expected second report output to be '25', got '%s'", reports[1].Stdout)
}
