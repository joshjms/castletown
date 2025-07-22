package exec

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshjms/castletown/sandbox"
)

func createFiles(filesDir string, files []File) error {
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return err
	}

	for _, f := range files {
		file, err := os.Create(filepath.Join(filesDir, f.Name))
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.WriteString(file, f.Content); err != nil {
			return err
		}
	}

	return nil
}

func getRootfsDir(imageDir string, image string) (string, error) {
	image = strings.Replace(image, ":", "-", 1)
	rootfsDir := filepath.Join(imageDir, image)
	if _, err := os.Stat(rootfsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("rootfs directory does not exist: %s", rootfsDir)
	}
	return rootfsDir, nil
}

func getCopyFiles(filesDir string, files []string) []sandbox.File {
	copyFiles := make([]sandbox.File, len(files))
	for i, file := range files {
		copyFiles[i] = sandbox.File{
			Src: filepath.Join(filesDir, file),
			Dst: filepath.Join("/box", file),
		}
	}
	return copyFiles
}

func getSaveFiles(filesDir string, persist []string) []sandbox.File {
	saveFiles := make([]sandbox.File, len(persist))
	for i, file := range persist {
		saveFiles[i] = sandbox.File{
			Src: filepath.Join("/box", file),
			Dst: filepath.Join(filesDir, file),
		}
	}
	return saveFiles
}
