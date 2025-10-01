package sandbox

import (
	"fmt"

	"github.com/containerd/cgroups/v3/cgroup2"
)

func loadCgroup(id string) (*cgroup2.Manager, error) {
	slicePath, err := getSlicePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get full slice path: %w", err)
	}

	mgr, err := cgroup2.LoadSystemd(slicePath, fmt.Sprintf("castletown-%s.scope", id))
	if err != nil {
		return nil, fmt.Errorf("failed to load cgroup manager: %w", err)
	}

	return mgr, nil
}

func getSlicePath() (string, error) {
	return "/castletown.slice", nil
}
