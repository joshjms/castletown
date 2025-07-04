package castletown

import (
	"bytes"
	"context"
	"fmt"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/specconv"
)

type Sandbox struct {
	ID     string
	Rootfs string
	Config *Config
}

func NewSandbox(id string, cfg *Config) *Sandbox {
	return &Sandbox{
		ID:     id,
		Config: cfg,
	}
}

func (s *Sandbox) Run(ctx context.Context) (*Report, error) {
	overlayfs, err := prepareRootfs(s.ID, s.Config.Rootfs, s.Config.UserNamespace)
	if err != nil {
		return nil, fmt.Errorf("error preparing rootfs: %w", err)
	}

	spec, err := createSpec(s.ID, s.Config, overlayfs)
	if err != nil {
		return nil, fmt.Errorf("error creating oci spec: %w", err)
	}

	libcontainerConfig, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		UseSystemdCgroup: false,
		Spec:             spec,
		RootlessEUID:     true,
		RootlessCgroups:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating libcontainer config: %w", err)
	}

	container, err := libcontainer.Create(LIBCONTAINER_ROOT, s.ID, libcontainerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating container: %w", err)
	}
	defer container.Destroy()

	noNewPrivileges := true

	var stdinBuf, stdoutBuf, stderrBuf bytes.Buffer

	if s.Config.Stdin != "" {
		stdinBuf.WriteString(s.Config.Stdin)
	}

	process := &libcontainer.Process{
		Args:            s.Config.Args,
		Env:             s.Config.Env,
		UID:             65534,
		GID:             65534,
		Cwd:             s.Config.Cwd,
		NoNewPrivileges: &noNewPrivileges,
		Stdin:           &stdinBuf,
		Stdout:          &stdoutBuf,
		Stderr:          &stderrBuf,
		Init:            true,
	}

	if err := container.Run(process); err != nil {
		return nil, fmt.Errorf("error running container: %w", err)
	}

	state, err := process.Wait()
	if err != nil {
		return nil, fmt.Errorf("error waiting for process: %w", err)
	}

	cgManager, err := loadCgroup(s.ID)
	if err != nil {
		return nil, fmt.Errorf("error loading cgroup: %w", err)
	}

	stats, err := cgManager.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting cgroup stats: %w", err)
	}

	report, err := MakeReport(&stdoutBuf, &stderrBuf, state, stats)
	if err != nil {
		return nil, fmt.Errorf("error making report: %w", err)
	}

	return report, nil
}
