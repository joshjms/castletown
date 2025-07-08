package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func createRootfs(image string) (string, error) {
	// Pull docker image as OCI image layout format

	sourceImage := fmt.Sprintf("docker://%s", image)

	ociImagePath := filepath.Join("/tmp", fmt.Sprintf("%s_%s", uuid.New().String(), image))
	destinationImage := fmt.Sprintf("oci:%s", ociImagePath)

	if err := exec.Command("skopeo", "copy", sourceImage, destinationImage).Run(); err != nil {
		return "", err
	}
	defer os.RemoveAll(strings.Split(ociImagePath, ":")[0])

	rootfsDir := filepath.Join("/tmp", fmt.Sprintf("_%d_%s", time.Now().Unix(), strings.Split(image, ":")[0]))

	// Unpack OCI image into rootfs

	if err := exec.Command("umoci", "raw", "unpack", "--rootless", "--image", ociImagePath, rootfsDir).Run(); err != nil {
		return "", err
	}

	return rootfsDir, nil
}
