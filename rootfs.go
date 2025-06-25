package castletown

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

func prepareRootfs(id string, rootfs string, userCfg *UserNamespaceConfig) (*Overlayfs, error) {
	lowerDir := rootfs

	sandboxRoot := filepath.Join(LAKESIDE_ROOT, id)

	upperDir := filepath.Join(sandboxRoot, "upper")
	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return nil, fmt.Errorf("error mkdir upperdir: %w", err)
	}

	workDir := filepath.Join(sandboxRoot, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("error mkdir workdir: %w", err)
	}

	if err := os.Chown(upperDir, int(userCfg.RootUID), int(userCfg.RootGID)); err != nil {
		return nil, fmt.Errorf("error chown upperdir: %w", err)
	}

	if err := os.Chown(workDir, int(userCfg.RootUID), int(userCfg.RootGID)); err != nil {
		return nil, fmt.Errorf("error chown workdir: %w", err)
	}

	return &Overlayfs{
		LowerDir: lowerDir,
		UpperDir: upperDir,
		WorkDir:  workDir,
	}, nil
}
