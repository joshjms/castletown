# Getting Started

This guide will get you started on how to set up the castletown server. The following procedure is run in an Ubuntu 25.04 LTS amd64 brand new GCP Compute Engine instance.

## Downloading `castletown`

```shell
curl -LO https://github.com/joshjms/castletown/releases/download/v0.2.0/castletown-linux-amd64
chmod +x castletown-linux-amd64
sudo mv castletown-linux-amd64 /usr/local/bin/castletown
castletown version
```

```
castletown v0.2.0
```

## Enabling Unprivileged User Namespaces

There might be some kernel parameters or apparmor configs that disallow user namespaces for unprivileged users. For `castletown` to work, we need to disable them.

### Kernel

```shell
sudo sysctl kernel.unprivileged_userns_clone=1
```

### AppArmor

```shell
sudo sysctl kernel.apparmor_restrict_unprivileged_userns=0
```

## Delegate cgroup Control to User Services

This step ensures that user services can manage their own cgroups. See [cgroupv2](https://docs.kernel.org/admin-guide/cgroup-v2.html#delegation) for more details.

```shell
sudo mkdir -p /etc/systemd/system/user@.service.d
cat <<EOF | sudo tee /etc/systemd/system/user@.service.d/delegate.conf
[Service]
Delegate=yes
EOF
```

## Creating Rootfs Directory

For this example, let's use the `gcc:15-bookworm` image. There are several ways to do this, the following seems simple enough.

### Install `skopeo` and `umoci`

#### Skopeo

[here](https://github.com/containers/skopeo)

#### Umoci

[here](https://umo.ci/)

### Copy container from Docker registry to a local OCI format layout

```shell
skopeo copy docker://gcc:15-bookworm oci:/tmp/_tmp_gcc:15-bookworm
```

### Unpack OCI image layout into a rootless container bundle

```shell
mkdir /home/$USER/images
umoci raw unpack --rootless \
    --image /tmp/_tmp_gcc:15-bookworm \
    /home/$USER/images/gcc-15-bookworm
```

## Adding `subuid` and `subgid`

```shell
sudo usermod --add-subuids 100000-165535 $USER
sudo usermod --add-subgids 100000-165535 $USER
```

## Running `castletown`

```shell
systemd-run --user --scope castletown server --images_dir=/home/$USER/images
```

```shell
Running as unit: run-p7706-i7707.scope; invocation ID: c34683ed66624d28808b5d066392317d
Starting server at port :8000
```

## Done!

Try sending a POST request to port `8000` with the following body.

```json
{
    "files": [
        {
            "name": "main.cpp",
            "content": "#include <iostream>\nint main() {\n  std::cout << \"Don't forget!\" << std::endl;\n  return 0;\n}"
        }
    ],
    "steps": [
        {
            "image": "gcc:15-bookworm",
            "cmd": [
                "g++",
                "main.cpp",
                "-o",
                "main"
            ],
            "stdin": "",
            "memoryLimitMB": 128,
            "timeLimitMs": 2000,
            "procLimit": 10,
            "files": [
                "main.cpp"
            ],
            "persist": [
                "main"
            ]
        },
        {
            "image": "gcc:15-bookworm",
            "cmd": [
                "./main"
            ],
            "stdin": "",
            "memoryLimitMB": 128,
            "timeLimitMs": 1000,
            "procLimit": 10,
            "files": [
                "main"
            ],
            "persist": []
        }
    ]
}
```

```json
[
    {
        "Status": "OK",
        "ExitCode": 0,
        "Signal": -1,
        "Stdout": "",
        "Stderr": "",
        "CPUTime": 553343,
        "Memory": 47112192,
        "WallTime": 0
    },
    {
        "Status": "OK",
        "ExitCode": 0,
        "Signal": -1,
        "Stdout": "Don't forget!\n",
        "Stderr": "",
        "CPUTime": 28809,
        "Memory": 2576384,
        "WallTime": 0
    }
]
```
