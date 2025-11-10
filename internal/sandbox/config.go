package sandbox

type Config struct {
	RootfsImageDir string

	Args  []string
	Stdin string
	Cwd   string
	Env   []string

	UserNamespace *UserNamespaceConfig

	TimeLimitMs int64
	Cgroup      *CgroupConfig
	Rlimit      *RlimitConfig

	BoxDir string
	Files  []File
}

type UserNamespaceConfig struct {
	HostUID      uint32
	HostGID      uint32
	ContainerUID uint32
	ContainerGID uint32
	UIDMapCount  uint32
	GIDMapCount  uint32
}

type CgroupConfig struct {
	CpuShares  uint64
	CpuQuota   int64
	CpusetCpus string
	CpusetMems string
	Memory     int64
	PidsLimit  int64
}

type RlimitConfig struct {
	Core   *Rlimit
	Fsize  *Rlimit
	NoFile *Rlimit
}

type Rlimit struct {
	Hard uint64
	Soft uint64
}

type File struct {
	Src     string
	Content string
	Dst     string
}
