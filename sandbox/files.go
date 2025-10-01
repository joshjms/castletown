package sandbox

import (
	"os"
	"path/filepath"
)

func (s *Sandbox) prepareFiles() error {
	for _, file := range s.config.Files {
		if err := os.MkdirAll(filepath.Dir(file.Dst), 0755); err != nil {
			return err
		}

		if file.Src != "" {
			err := copyFile(file.Src, file.Dst)
			if err != nil {
				return err
			}
		} else {
			err := writeFile(file.Content, file.Dst)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0744)
}

func writeFile(content, dst string) error {
	return os.WriteFile(dst, []byte(content), 0744)
}
