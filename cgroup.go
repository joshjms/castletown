package castletown

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/coreos/go-systemd/v22/dbus"
)

func loadCgroup(id string) (*cgroup2.Manager, error) {
	slicePath, err := getSlicePath(os.Getuid())
	if err != nil {
		return nil, fmt.Errorf("failed to get full slice path: %w", err)
	}

	mgr, err := cgroup2.LoadSystemd(slicePath, fmt.Sprintf("castletown-%s.scope", id))
	if err != nil {
		return nil, fmt.Errorf("failed to load cgroup manager: %w", err)
	}

	return mgr, nil
}

func getSlicePath(uid int) (string, error) {
	service := fmt.Sprintf("user@%d.service", uid)

	ctx := context.Background()

	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return "", fmt.Errorf("error connecting to systemd: %w", err)
	}
	defer conn.Close()

	prop, err := conn.GetServicePropertyContext(ctx, service, "ControlGroup")
	if err != nil {
		return "", fmt.Errorf("error getting ControlGroup property: %w", err)
	}

	slicePath, ok := prop.Value.Value().(string)
	if !ok {
		return "", fmt.Errorf("ControlGroup property is not a string")
	}
	if slicePath == "" {
		return "", fmt.Errorf("ControlGroup property is empty")
	}

	return slicePath, nil
}
