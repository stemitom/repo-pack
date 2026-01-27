package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ProgressCallback func(bytesWritten int64)

type progressWriter struct {
	writer   io.Writer
	callback ProgressCallback
	written  int64
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		pw.written += int64(n)
		if pw.callback != nil {
			pw.callback(pw.written)
		}
	}
	return n, err
}

func absPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(path)
}

func FileExists(baseDir string, filePath string, outputDir string) (bool, error) {
	adjustedFilePath, err := extractRelativePath(baseDir, filePath)
	if err != nil {
		return false, err
	}

	fullPath := filepath.Join(outputDir, adjustedFilePath)
	fullPath = filepath.Clean(fullPath)

	absOutputDir, err := absPath(outputDir)
	if err != nil {
		return false, fmt.Errorf("error resolving output directory: %w", err)
	}
	absFullPath, err := absPath(fullPath)
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

func FileExistsDirect(filePath string, outputDir string) (bool, error) {
	fullPath := filepath.Join(outputDir, filePath)
	fullPath = filepath.Clean(fullPath)

	absOutputDir, err := absPath(outputDir)
	if err != nil {
		return false, fmt.Errorf("error resolving output directory: %w", err)
	}
	absFullPath, err := absPath(fullPath)
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

func SaveFile(baseDir string, filePath string, reader io.ReadCloser, outputDir string) error {
	return SaveFileWithProgress(baseDir, filePath, reader, outputDir, nil)
}

func SaveFileWithProgress(baseDir string, filePath string, reader io.ReadCloser, outputDir string, onProgress ProgressCallback) error {
	defer reader.Close()

	adjustedFilePath, err := extractRelativePath(baseDir, filePath)
	if err != nil {
		return err
	}

	fullPath := filepath.Join(outputDir, adjustedFilePath)
	fullPath = filepath.Clean(fullPath)

	absOutputDir, err := absPath(outputDir)
	if err != nil {
		return fmt.Errorf("error resolving output directory: %w", err)
	}
	absFullPath, err := absPath(fullPath)
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

	var writer io.Writer = file
	if onProgress != nil {
		writer = &progressWriter{writer: file, callback: onProgress}
	}

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("error copying content to file %s: %w", fullPath, err)
	}

	return nil
}

func SaveFileDirect(filePath string, reader io.ReadCloser, outputDir string, onProgress ProgressCallback) error {
	defer reader.Close()

	fullPath := filepath.Join(outputDir, filePath)
	fullPath = filepath.Clean(fullPath)

	absOutputDir, err := absPath(outputDir)
	if err != nil {
		return fmt.Errorf("error resolving output directory: %w", err)
	}
	absFullPath, err := absPath(fullPath)
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

	var writer io.Writer = file
	if onProgress != nil {
		writer = &progressWriter{writer: file, callback: onProgress}
	}

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("error copying content to file %s: %w", fullPath, err)
	}

	return nil
}

func extractRelativePath(baseDir string, filePath string) (string, error) {
	baseDir = filepath.Clean(baseDir)
	filePath = filepath.Clean(filePath)

	baseDirIndex := strings.Index(filePath, baseDir+string(filepath.Separator))
	if baseDirIndex == -1 {
		if strings.HasSuffix(filePath, baseDir) {
			return "", nil
		}
		return "", fmt.Errorf("base directory %s not found in file path %s", baseDir, filePath)
	}

	return filePath[baseDirIndex:], nil
}
