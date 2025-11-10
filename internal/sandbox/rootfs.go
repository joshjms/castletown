package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Sandbox) prepareOverlayfs() error {
	upperDir := filepath.Join(s.runtimeCfg.OverlayFSDir, fmt.Sprintf("sandbox-%s", s.id), "upper")
	workDir := filepath.Join(s.runtimeCfg.OverlayFSDir, fmt.Sprintf("sandbox-%s", s.id), "work")

	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	return nil
}

func (s *Sandbox) getLowerDir() string {
	return s.config.RootfsImageDir
}

func (s *Sandbox) getUpperDir() string {
	return filepath.Join(s.runtimeCfg.OverlayFSDir, fmt.Sprintf("sandbox-%s", s.id), "upper")
}

func (s *Sandbox) getWorkDir() string {
	return filepath.Join(s.runtimeCfg.OverlayFSDir, fmt.Sprintf("sandbox-%s", s.id), "work")
}
