package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SaveFile saves file to a filepath and base directory
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
	if makeDirErr := os.MkdirAll(dir, 0o755); makeDirErr != nil && !os.IsExist(makeDirErr) {
		return fmt.Errorf("error creating output folder for %s: %w", fullPath, makeDirErr)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", fullPath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("error copying content to file %s: %v", fullPath, err)
	}

	return nil
}
