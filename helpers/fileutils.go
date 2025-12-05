package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileExists checks if a file exists at the given path
func FileExists(baseDir string, filePath string, outputDir string) (bool, error) {
	adjustedFilePath, err := extractRelativePath(baseDir, filePath)
	if err != nil {
		return false, err
	}

	fullPath := filepath.Join(outputDir, adjustedFilePath)
	fullPath = filepath.Clean(fullPath)

	// Ensure fullPath is within outputDir
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return false, fmt.Errorf("error resolving output directory: %w", err)
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false, fmt.Errorf("error resolving file path: %w", err)
	}
	if !strings.HasPrefix(absFullPath, absOutputDir+string(filepath.Separator)) && absFullPath != absOutputDir {
		return false, fmt.Errorf("%s is outside output directory %s", filePath, outputDir)
	}

	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// SaveFile saves file to a filepath and base directory
func SaveFile(baseDir string, filePath string, reader io.ReadCloser, outputDir string) error {
	defer reader.Close()

	adjustedFilePath, err := extractRelativePath(baseDir, filePath)
	if err != nil {
		return err
	}

	fullPath := filepath.Join(outputDir, adjustedFilePath)
	fullPath = filepath.Clean(fullPath)

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("error resolving output directory: %w", err)
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("error resolving file path: %w", err)
	}
	if !strings.HasPrefix(absFullPath, absOutputDir+string(filepath.Separator)) && absFullPath != absOutputDir {
		return fmt.Errorf("%s is outside output directory %s", filePath, outputDir)
	}

	dir := filepath.Dir(fullPath)
	if makeDirErr := os.MkdirAll(dir, 0o755); makeDirErr != nil && !os.IsExist(makeDirErr) {
		return fmt.Errorf("error creating output folder for %s: %w", fullPath, makeDirErr)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", fullPath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("error copying content to file %s: %w", fullPath, err)
	}

	return nil
}

// extractRelativePath extracts the relative path starting from baseDir
func extractRelativePath(baseDir string, filePath string) (string, error) {
	// Normalize both paths
	baseDir = filepath.Clean(baseDir)
	filePath = filepath.Clean(filePath)

	// Look for baseDir as a path component
	baseDirIndex := strings.Index(filePath, baseDir+string(filepath.Separator))
	if baseDirIndex == -1 {
		// Try without separator at the end for exact match at end
		if strings.HasSuffix(filePath, baseDir) {
			return "", nil
		}
		return "", fmt.Errorf("base directory %s not found in file path %s", baseDir, filePath)
	}

	return filePath[baseDirIndex:], nil
}
