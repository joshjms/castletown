package sandbox

func GetDefaultConfig() *Config {
	return &Config{
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		Cwd:         "/box",
		TimeLimitMs: 1000,
		Cgroup: &CgroupConfig{
			CpuShares:  100000,
			CpuQuota:   100000,
			Memory:     256 * 1024 * 1024,
			PidsLimit:  100,
			CpusetCpus: "0",
			CpusetMems: "0",
		},
		Rlimit: &RlimitConfig{
			Core: &Rlimit{
				Hard: 0,
				Soft: 0,
			},
			Fsize: &Rlimit{
				Hard: 1 * 1024 * 1024,
				Soft: 1 * 1024 * 1024,
			},
			NoFile: &Rlimit{
				Hard: 64,
				Soft: 64,
			},
		},
	}
}
