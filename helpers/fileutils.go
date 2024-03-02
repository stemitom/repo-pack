package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SaveFile(baseDir string, filePath string, content []byte) error {
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

	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return fmt.Errorf("error saving file %s: %w", fullPath, err)
	}

	return nil
}
