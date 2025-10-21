package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joshjms/castletown/config"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"golang.org/x/sys/unix"
)

type Sandbox struct {
	id     string
	config *Config

	container *libcontainer.Container
}

func (s *Sandbox) GetId() string {
	return s.id
}

// Run runs a command inside the sandbox and returns a Report
func (s *Sandbox) Run(ctx context.Context) (Report, error) {
	err := s.prepareOverlayfs()
	if err != nil {
		return Report{}, fmt.Errorf("error preparing rootfs: %w", err)
	}

	if err := s.prepareFiles(); err != nil {
		return Report{}, fmt.Errorf("error preparing files: %w", err)
	}

	spec, err := s.createSpec()
	if err != nil {
		return Report{}, fmt.Errorf("error creating oci spec: %w", err)
	}

	libcontainerConfig, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		UseSystemdCgroup: false,
		Spec:             spec,
		// RootlessEUID:     true,
		// RootlessCgroups:  true,
	})
	if err != nil {
		return Report{}, fmt.Errorf("error creating libcontainer config: %w", err)
	}

	container, err := libcontainer.Create(config.LibcontainerDir, s.id, libcontainerConfig)
	if err != nil {
		return Report{}, fmt.Errorf("error creating container: %w", err)
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
		UID:             0,
		GID:             0,
		Cwd:             s.config.Cwd,
		NoNewPrivileges: &noNewPrivileges,
		Stdin:           &stdinBuf,
		Stdout:          &stdoutBuf,
		Stderr:          &stderrBuf,
		Rlimits:         getRlimits(s.config.Rlimit),
		Init:            true,
	}

	startAt := time.Now()

	if err := container.Run(process); err != nil {
		return Report{}, fmt.Errorf("error running container: %w", err)
	}

	processFinished := make(chan interface{}, 1)
	timeLimitExceeded := false

	go func() {
		select {
		case <-processFinished:
		case <-time.After(time.Duration(s.config.TimeLimitMs) * time.Millisecond * 3):
			timeLimitExceeded = true
			container.Signal(unix.SIGKILL)
		}
	}()

	state, _ := process.Wait()
	processFinished <- struct{}{}

	finishAt := time.Now()

	return s.makeReport(&stdoutBuf, &stderrBuf, state, timeLimitExceeded, startAt, finishAt)
}

func getRlimits(cfg *RlimitConfig) []configs.Rlimit {
	if cfg == nil {
		return nil
	}

	var rlimits []configs.Rlimit

	if cfg.Core != nil {
		rlimits = append(rlimits, configs.Rlimit{
			Type: unix.RLIMIT_CORE,
			Hard: cfg.Core.Hard,
			Soft: cfg.Core.Soft,
		})
	}

	if cfg.Fsize != nil {
		rlimits = append(rlimits, configs.Rlimit{
			Type: unix.RLIMIT_FSIZE,
			Hard: cfg.Fsize.Hard,
			Soft: cfg.Fsize.Soft,
		})
	}

	if cfg.NoFile != nil {
		rlimits = append(rlimits, configs.Rlimit{
			Type: unix.RLIMIT_NOFILE,
			Hard: cfg.NoFile.Hard,
			Soft: cfg.NoFile.Soft,
		})
	}

	return rlimits
}

func (s *Sandbox) Destroy() error {
	if s.container != nil {
		s.container.Destroy()
	}

	if err := os.RemoveAll(filepath.Join(config.OverlayFSDir, fmt.Sprintf("sandbox-%s", s.id))); err != nil {
		return fmt.Errorf("error removing overlayfs dirs: %w", err)
	}

	return nil
}
