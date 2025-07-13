package sandbox

type Config struct {
	RootfsImageDir string

	Args  []string
	Stdin string
	Cwd   string
	Env   []string

	ContainerUID  int
	ContainerGID  int
	UserNamespace *UserNamespaceConfig

	TimeLimitMs int64
	Cgroup      *CgroupConfig
	Rlimit      *RlimitConfig

	Copy []File
	Save []File
}

type UserNamespaceConfig struct {
	RootUID     uint32
	UIDMapStart uint32
	UIDMapCount uint32
	RootGID     uint32
	GIDMapStart uint32
	GIDMapCount uint32
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
	Src string
	Dst string
}
