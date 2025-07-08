package sandbox

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opencontainers/runtime-spec/specs-go"

	_ "github.com/opencontainers/cgroups/devices"
)

func createSpec(id string, cfg *Config, overlayfs *Overlayfs) (*specs.Spec, error) {
	slicePath, err := getSlicePath(os.Getuid())
	if err != nil {
		return nil, fmt.Errorf("failed to get slice path: %w", err)
	}

	rootfsMount := specs.Mount{
		Destination: "/",
		Type:        "overlay",
		Source:      "overlay",
		Options: []string{
			"rw",
			"userxattr",
			"xino=off",
			"index=off",
			fmt.Sprintf("upperdir=%s", overlayfs.UpperDir),
			fmt.Sprintf("lowerdir=%s", overlayfs.LowerDir),
			fmt.Sprintf("workdir=%s", overlayfs.WorkDir),
		},
	}

	mounts := []specs.Mount{}
	mounts = append(mounts, rootfsMount)
	mounts = append(mounts, defaultMounts()...)

	spec := &specs.Spec{
		Version: specs.Version,
		Process: &specs.Process{
			NoNewPrivileges: true,
		},
		Root: &specs.Root{
			Path:     "/home/joshjms/bundle/rootfs",
			Readonly: false,
		},
		Hostname: "castletown",
		Mounts:   mounts,
		Linux: &specs.Linux{
			CgroupsPath: filepath.Join(slicePath, fmt.Sprintf("castletown-%s.scope", id), id),
			Resources:   cgroupResources(cfg.Cgroup),
			UIDMappings: uidMappings(cfg.UserNamespace),
			GIDMappings: gidMappings(cfg.UserNamespace),
			Namespaces: []specs.LinuxNamespace{
				{
					Type: specs.CgroupNamespace,
				},
				{
					Type: specs.PIDNamespace,
				},
				{
					Type: specs.IPCNamespace,
				},
				{
					Type: specs.UTSNamespace,
				},
				{
					Type: specs.MountNamespace,
				},
				{
					Type: specs.UserNamespace,
				},
				{
					Type: specs.NetworkNamespace,
				},
			},
			// https://github.com/moby/moby/blob/master/oci/defaults.go
			MaskedPaths: []string{
				"/proc/asound",
				"/proc/acpi",
				"/proc/interrupts", // https://github.com/moby/moby/security/advisories/GHSA-6fw5-f8r9-fgfm
				"/proc/kcore",
				"/proc/keys",
				"/proc/latency_stats",
				"/proc/timer_list",
				"/proc/timer_stats",
				"/proc/sched_debug",
				"/proc/scsi",
				"/sys/firmware",
				"/sys/devices/virtual/powercap", // https://github.com/moby/moby/security/advisories/GHSA-jq35-85cj-fj4p
			},
			ReadonlyPaths: []string{
				"/proc/bus",
				"/proc/fs",
				"/proc/irq",
				"/proc/sys",
				"/proc/sysrq-trigger",
			},
		},
	}

	return spec, nil
}

func defaultMounts() []specs.Mount {
	return []specs.Mount{
		{
			Destination: "/proc",
			Type:        "proc",
			Source:      "proc",
		},
		{
			Destination: "/dev",
			Type:        "tmpfs",
			Source:      "tmpfs",
			Options: []string{
				"nosuid",
				"strictatime",
				"mode=755",
				"size=65536k",
			},
		},
		{
			Destination: "/dev/pts",
			Type:        "devpts",
			Source:      "devpts",
			Options: []string{
				"nosuid",
				"noexec",
				"newinstance",
				"ptmxmode=0666",
				"mode=0620",
			},
		},
		{
			Destination: "/dev/shm",
			Type:        "tmpfs",
			Source:      "shm",
			Options: []string{
				"nosuid",
				"noexec",
				"nodev",
				"mode=1777",
				"size=65536k",
			},
		},
		{
			Destination: "/dev/mqueue",
			Type:        "mqueue",
			Source:      "mqueue",
			Options: []string{
				"nosuid",
				"noexec",
				"nodev",
			},
		},
		{
			Destination: "/tmp",
			Type:        "tmpfs",
			Source:      "tmpfs",
			Options: []string{
				"nosuid",
				"noexec",
				"nodev",
				"size=128m",
				"nr_inodes=4k",
			},
		},
	}
}

func cgroupResources(cfg *CgroupConfig) *specs.LinuxResources {
	cgMemory := &specs.LinuxMemory{}
	cgCPU := &specs.LinuxCPU{}
	cgPids := &specs.LinuxPids{}

	if cfg.CpuQuota != 0 {
		cgCPU.Quota = &cfg.CpuQuota
	}

	if cfg.CpuShares != 0 {
		cgCPU.Shares = &cfg.CpuShares
	}

	if cfg.CpusetCpus != "" {
		cgCPU.Cpus = cfg.CpusetCpus
	}

	if cfg.CpusetMems != "" {
		cgCPU.Mems = cfg.CpusetMems
	}

	if cfg.Memory != 0 {
		cgMemory.Limit = &cfg.Memory
		cgMemory.Swap = &cfg.Memory
	}

	if cfg.PidsLimit != 0 {
		cgPids.Limit = cfg.PidsLimit
	}

	return &specs.LinuxResources{
		CPU:    cgCPU,
		Memory: cgMemory,
		Pids:   cgPids,
	}
}

func uidMappings(cfg *UserNamespaceConfig) []specs.LinuxIDMapping {
	return []specs.LinuxIDMapping{
		{
			ContainerID: 0,
			HostID:      cfg.RootUID,
			Size:        1,
		},
		{
			ContainerID: 1,
			HostID:      cfg.UIDMapStart,
			Size:        cfg.UIDMapCount,
		},
	}
}

func gidMappings(cfg *UserNamespaceConfig) []specs.LinuxIDMapping {
	return []specs.LinuxIDMapping{
		{
			ContainerID: 0,
			HostID:      cfg.RootUID,
			Size:        1,
		},
		{
			ContainerID: 1,
			HostID:      cfg.GIDMapStart,
			Size:        cfg.GIDMapCount,
		},
	}
}
