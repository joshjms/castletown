# castletown

## Installation

```sh
go get github.com/joshjms/castletown
```

## Usage

```go
castletown.Init()

config := &castletown.Config{
	Rootfs: "/var/lib/sandbox/images/ubuntu",
	Args:   []string{"sh", "-c", "echo \"don't forget\""},
	Cwd:    "/",
	Env: []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	},
	UserNamespace: &castletown.UserNamespaceConfig{
		RootUID:     1000,
		UIDMapStart: 100000,
		UIDMapCount: 65534,
		RootGID:     1000,
		GIDMapStart: 100000,
		GIDMapCount: 65534,
	},
	Cgroup: &castletown.CgroupConfig{
		CpuQuota: 1000000,
		Memory:   512 * 1024 * 1024,
	},
}

sandbox := castletown.NewSandbox("sandbox-01", config)
report, err := sandbox.Run(context.Background())
if err != nil {
	panic(err)
}

fmt.Println(report.Stdout)
```

#### Output

```
don't forget
```
