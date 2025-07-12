package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

type Sandbox struct {
	id     string
	config *Config

	spec      *specs.Spec
	overlayfs *Overlayfs
}

func NewSandbox(id string, cfg *Config) *Sandbox {
	return &Sandbox{
		id:     id,
		config: cfg,
	}
}

// Run runs a command inside the sandbox and returns a Report
func (s *Sandbox) Run(ctx context.Context) (*Report, error) {
	overlayfs, err := s.prepare()
	if err != nil {
		return nil, fmt.Errorf("error preparing rootfs: %w", err)
	}
	defer s.destroy()
	s.overlayfs = overlayfs

	if err := s.copy(); err != nil {
		return nil, fmt.Errorf("error copying files into sandbox: %w", err)
	}
	defer s.save()

	spec, err := createSpec(s.id, s.config, overlayfs)
	if err != nil {
		return nil, fmt.Errorf("error creating oci spec: %w", err)
	}
	s.spec = spec

	libcontainerConfig, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		UseSystemdCgroup: false,
		Spec:             s.spec,
		RootlessEUID:     true,
		RootlessCgroups:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating libcontainer config: %w", err)
	}

	container, err := libcontainer.Create(LIBCONTAINER_ROOT, s.id, libcontainerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating container: %w", err)
	}
	defer container.Destroy()

	noNewPrivileges := true

	var stdinBuf, stdoutBuf, stderrBuf bytes.Buffer

	if s.config.Stdin != "" {
		stdinBuf.WriteString(s.config.Stdin)
	}

	process := &libcontainer.Process{
		Args:            s.config.Args,
		Env:             s.config.Env,
		UID:             s.config.ContainerUID,
		GID:             s.config.ContainerGID,
		Cwd:             s.config.Cwd,
		NoNewPrivileges: &noNewPrivileges,
		Stdin:           &stdinBuf,
		Stdout:          &stdoutBuf,
		Stderr:          &stderrBuf,
		Init:            true,
	}

	if err := container.Run(process); err != nil {
		return nil, fmt.Errorf("error running container: %w", err)
	}

	processFinished := make(chan interface{}, 1)
	timeLimitExceeded := false

	go func() {
		select {
		case <-processFinished:
		case <-time.After(time.Duration(s.config.TimeLimitMs) * time.Millisecond):
			timeLimitExceeded = true
			container.Signal(unix.SIGKILL)
		}
	}()

	state, _ := process.Wait()
	processFinished <- struct{}{}

	return s.makeReport(&stdoutBuf, &stderrBuf, state, timeLimitExceeded)
}
