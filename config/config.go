package config

var (
	OverlayFSDir    string
	StorageDir      string
	ImagesDir       string
	LibcontainerDir string
	RootfsDir       string

	MaxConcurrency int
	Port           int
)

func UseDefaults() {
	OverlayFSDir = "/tmp/castletown/overlayfs"
	StorageDir = "/tmp/castletown/storage"
	ImagesDir = "/tmp/castletown/images"
	LibcontainerDir = "/tmp/castletown/libcontainer"
	RootfsDir = "/tmp/castletown/rootfs"

	MaxConcurrency = 10
	Port = 8080
}
