package sandbox

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type Testcase struct {
	File  string
	Stdin string

	ExpectedStatus *Status
	ExpectedOutput *string

	TimeLimitMs int64
}

func (tc *Testcase) Run() error {
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
	if err != nil {
		return err
	}
	if compileReport.Status != STATUS_OK {
		return fmt.Errorf("status not ok")
	}

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
			Memory:    128 * 1024 * 1024,
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
	if err != nil {
		return err
	}

	if tc.ExpectedStatus != nil && execReport.Status != *tc.ExpectedStatus {
		return fmt.Errorf("status != expectedStatus, expected: %v, got: %v", *tc.ExpectedStatus, execReport.Status)
	}

	if tc.ExpectedOutput != nil && execReport.Stdout != *tc.ExpectedOutput {
		return fmt.Errorf("output != expectedOutput, expected: %v, got: %v", *tc.ExpectedOutput, execReport.Stdout)
	}

	return nil
}
