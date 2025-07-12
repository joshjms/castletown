package sandbox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (s *Sandbox) copy() error {
	files := s.config.Copy

	for _, f := range files {
		dstPath := filepath.Join(s.overlayfs.UpperDir, f.Dst)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("error creating directories: %w", err)
		}

		srcInfo, err := os.Stat(f.Src)
		if err != nil {
			return fmt.Errorf("error src file does not exist: %w", err)
		}
		srcMode := srcInfo.Mode()

		srcFile, err := os.Open(f.Src)
		if err != nil {
			return fmt.Errorf("error opening src file: %w", err)
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcMode.Perm())
		if err != nil {
			return fmt.Errorf("error creating dst file: %w", err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("error copying src to dst: %w", err)
		}

		if err := os.Chown(dstPath,
			int(s.config.UserNamespace.RootUID),
			int(s.config.UserNamespace.RootGID),
		); err != nil {
			return fmt.Errorf("error chown file: %w", err)
		}
	}
	return nil
}

func (s *Sandbox) save() error {
	files := s.config.Save

	for _, f := range files {
		srcPath := filepath.Join(s.overlayfs.UpperDir, f.Src)
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("error src file does not exist: %w", err)
		}
		srcMode := srcInfo.Mode()

		srcFile, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("error opening src file: %w", err)
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(f.Dst), 0755); err != nil {
			return fmt.Errorf("error creating dir: %w", err)
		}

		dstFile, err := os.OpenFile(f.Dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcMode.Perm())
		if err != nil {
			return fmt.Errorf("error creating dst file: %w", err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("error copying src to dst: %w", err)
		}
	}
	return nil
}
