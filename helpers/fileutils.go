package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func SaveFile(baseDir string, filePath string, reader io.ReadCloser) error {
	defer reader.Close()
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}

	baseDirIndex := strings.Index(filePath, baseDir+"/")
	if baseDirIndex == -1 {
		return fmt.Errorf("base directory %s not found in file path %s", baseDir, filePath)
	}

	adjustedFilePath := filePath[baseDirIndex:]
	fullPath := filepath.Join(currentDir, adjustedFilePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating output folder for %s: %w", fullPath, err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", fullPath, err)
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("error copying content to file %s: %v", fullPath, err)
	}

	defer file.Close()
	return nil
}
