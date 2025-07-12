package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
)

type Overlayfs struct {
	LowerDir string
	UpperDir string
	WorkDir  string
}

func (s *Sandbox) prepare() (*Overlayfs, error) {
	lowerDir := s.config.RootfsImageDir

	sandboxRoot := filepath.Join(CONTAINERS_ROOT, s.id)

	upperDir := filepath.Join(sandboxRoot, "upper")
	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return nil, fmt.Errorf("error mkdir upperdir: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(upperDir, "box"), 0777); err != nil {
		return nil, fmt.Errorf("error mkdir writeable box directory: %w", err)
	}

	workDir := filepath.Join(sandboxRoot, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("error mkdir workdir: %w", err)
	}

	if err := os.Chown(upperDir, int(s.config.UserNamespace.RootUID), int(s.config.UserNamespace.RootGID)); err != nil {
		return nil, fmt.Errorf("error chown upperdir: %w", err)
	}

	if err := os.Chown(workDir, int(s.config.UserNamespace.RootUID), int(s.config.UserNamespace.RootGID)); err != nil {
		return nil, fmt.Errorf("error chown workdir: %w", err)
	}

	return &Overlayfs{
		LowerDir: lowerDir,
		UpperDir: upperDir,
		WorkDir:  workDir,
	}, nil
}
func (s *Sandbox) destroy() error {
	sandboxRoot := filepath.Join(CONTAINERS_ROOT, s.id)

	if err := os.RemoveAll(sandboxRoot); err != nil {
		return err
	}

	return nil
}
