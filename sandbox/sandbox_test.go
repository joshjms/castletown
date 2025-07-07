package sandbox

import (
	"context"
	"testing"
)

func TestSandbox(t *testing.T) {
	Init()

	config := &Config{
		Rootfs: "/var/lib/sandbox/images/ubuntu",
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
	report, err := sandbox.Run(context.Background())
	if err != nil {
		t.Fatalf("failed to execute sandbox: %v", err)
	}

	if report.ExitCode != 0 || report.Stdout != "don't forget\n" {
		t.Errorf("unexpected report: %+v", report)
	}
	t.Logf("Sandbox executed successfully: %+v", report)
}
