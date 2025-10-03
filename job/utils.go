package job

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshjms/castletown/config"
	"github.com/joshjms/castletown/sandbox"
)

func getImageDir(image string) string {
	image = strings.Replace(image, ":", "-", 1)
	return filepath.Join(config.ImagesDir, image)
}

func verifyImages(procs []Process) error {
	for _, process := range procs {
		image := process.Image
		rootfsDir := getImageDir(image)

		f, err := os.Stat(rootfsDir)
		if os.IsNotExist(err) {
			return fmt.Errorf("rootfs directory does not exist: %s", rootfsDir)
		}
		if !f.IsDir() {
			return fmt.Errorf("rootfs path exists but is not a directory: %s", rootfsDir)
		}
	}

	return nil
}

func prepareFileDirs(reqId string, procs []Process) error {
	rootFileDir := filepath.Join(config.StorageDir, reqId)
	if err := os.MkdirAll(rootFileDir, 0755); err != nil {
		return fmt.Errorf("cannot create root files directory: %v", err)
	}

	for i := range procs {
		procDir := filepath.Join(rootFileDir, fmt.Sprintf("proc-%d", i))
		if err := os.MkdirAll(procDir, 0755); err != nil {
			return fmt.Errorf("cannot create process directory: %v", err)
		}
	}

	return nil
}

func getRootFileDir(reqId string) string {
	return filepath.Join(config.StorageDir, reqId)
}

func getProcFileDir(reqId string, procIndex int) string {
	return filepath.Join(getRootFileDir(reqId), fmt.Sprintf("proc-%d", procIndex))
}

func getFileDependencies(reqId string, procs []Process, files []File, step int) ([]sandbox.File, error) {
	fileMap := make(map[string]File)
	for _, file := range files {
		fileMap[file.Name] = file
	}

	fileDeps := make([]sandbox.File, 0)
	lastOcc := make(map[string]int)

	procDir := getProcFileDir(reqId, step)

	for i, proc := range procs[:step] {
		for _, fileName := range proc.Persist {
			lastOcc[fileName] = i
		}
	}

	proc := procs[step]

	for _, fileName := range proc.Files {
		if _, exists := lastOcc[fileName]; !exists {
			file, exists := fileMap[fileName]
			if !exists {
				return nil, fmt.Errorf("file %s not found", fileName)
			}

			fileDeps = append(fileDeps, sandbox.File{
				Content: file.Content,
				Dst:     filepath.Join(procDir, fileName),
			})
		} else {
			fileDeps = append(fileDeps, sandbox.File{
				Src: filepath.Join(getProcFileDir(reqId, lastOcc[fileName]), fileName),
				Dst: filepath.Join(procDir, fileName),
			})
		}
	}

	return fileDeps, nil
}
