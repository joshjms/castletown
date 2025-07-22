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

func (s *Sandbox) prepareOverlayfs() error {
	fsDir := filepath.Join(s.overlayfsDir, s.id)

	lowerDir := s.config.RootfsImageDir

	upperDir := filepath.Join(fsDir, "upper")
	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return fmt.Errorf("error mkdir upperdir: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(upperDir, "box"), 0777); err != nil {
		return fmt.Errorf("error mkdir writeable box directory: %w", err)
	}

	workDir := filepath.Join(fsDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("error mkdir workdir: %w", err)
	}

	if err := os.Chown(upperDir, int(s.config.UserNamespace.RootUID), int(s.config.UserNamespace.RootGID)); err != nil {
		return fmt.Errorf("error chown upperdir: %w", err)
	}

	if err := os.Chown(workDir, int(s.config.UserNamespace.RootUID), int(s.config.UserNamespace.RootGID)); err != nil {
		return fmt.Errorf("error chown workdir: %w", err)
	}

	s.overlayfs = &Overlayfs{
		LowerDir: lowerDir,
		UpperDir: upperDir,
		WorkDir:  workDir,
	}

	return nil
}

func (s *Sandbox) destroy() error {
	fsDir := filepath.Join(s.overlayfsDir, s.id)
	if err := os.RemoveAll(fsDir); err != nil {
		return err
	}
	return nil
}
