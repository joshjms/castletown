package sandbox

import (
	"fmt"
	"path/filepath"

	"github.com/opencontainers/runtime-spec/specs-go"

	_ "github.com/opencontainers/cgroups/devices"
)

func (s *Sandbox) createSpec() (*specs.Spec, error) {
	slicePath, err := getSlicePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get slice path: %w", err)
	}

	mounts := s.getMounts()

	spec := &specs.Spec{
		Version: specs.Version,
		Process: &specs.Process{
			NoNewPrivileges: true,
		},
		Root: &specs.Root{
			Path:     s.runtimeCfg.RootfsDir,
			Readonly: false,
		},
		Hostname: fmt.Sprintf("castletown-%s", s.id),
		Mounts:   mounts,
		Linux: &specs.Linux{
			CgroupsPath: filepath.Join(slicePath, fmt.Sprintf("castletown-%s.scope", s.id), s.id),
			Resources:   cgroupResources(s.config.Cgroup),
			UIDMappings: []specs.LinuxIDMapping{
				{
					HostID:      s.config.UserNamespace.HostUID,
					ContainerID: s.config.UserNamespace.ContainerUID,
					Size:        s.config.UserNamespace.UIDMapCount,
				},
			},
			GIDMappings: []specs.LinuxIDMapping{
				{
					HostID:      s.config.UserNamespace.HostGID,
					ContainerID: s.config.UserNamespace.ContainerGID,
					Size:        s.config.UserNamespace.GIDMapCount,
				},
			},
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

func (s *Sandbox) getMounts() []specs.Mount {
	mounts := make([]specs.Mount, 0)

	rootfsMount := specs.Mount{
		Destination: "/",
		Type:        "overlay",
		Source:      "overlay",
		Options: []string{
			"rw",
			"userxattr",
			"xino=off",
			"index=off",
			fmt.Sprintf("upperdir=%s", s.getUpperDir()),
			fmt.Sprintf("lowerdir=%s", s.getLowerDir()),
			fmt.Sprintf("workdir=%s", s.getWorkDir()),
		},
	}

	mounts = append(mounts, rootfsMount)

	bindMount := specs.Mount{
		Destination: "/box",
		Type:        "bind",
		Source:      s.config.BoxDir,
		Options: []string{
			"rbind",
			"rw",
			"exec",
			"nosuid",
			"nodev",
			"ridmap",
		},
		UIDMappings: []specs.LinuxIDMapping{
			{
				ContainerID: 0,
				HostID:      s.config.UserNamespace.HostUID,
				Size:        1,
			},
		},
		GIDMappings: []specs.LinuxIDMapping{
			{
				ContainerID: 0,
				HostID:      s.config.UserNamespace.HostGID,
				Size:        1,
			},
		},
	}

	mounts = append(mounts, bindMount)

	mounts = append(mounts, defaultMounts()...)

	return mounts
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
		{
			Destination: "/sys",
			Type:        "sysfs",
			Source:      "sysfs",
			Options:     []string{"nosuid", "noexec", "nodev", "ro"},
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
