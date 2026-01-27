package gh

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

type FileCache struct {
	cacheDir string
	enabled  bool
}

func NewFileCache() (*FileCache, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return &FileCache{enabled: false}, nil
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return &FileCache{enabled: false}, nil
	}

	return &FileCache{
		cacheDir: cacheDir,
		enabled:  true,
	}, nil
}

func getCacheDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(homeDir, "Library", "Caches")
	case "windows":
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			return "", fmt.Errorf("LOCALAPPDATA not set")
		}
	default:
		baseDir = os.Getenv("XDG_CACHE_HOME")
		if baseDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(homeDir, ".cache")
		}
	}

	return filepath.Join(baseDir, "repo-pack", "files"), nil
}

func (c *FileCache) Get(sha string, destPath string) (bool, error) {
	if !c.enabled || sha == "" {
		return false, nil
	}

	cachePath := c.cachePath(sha)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return false, err
	}

	if err := os.Link(cachePath, destPath); err == nil {
		return true, nil
	}

	return c.copyFile(cachePath, destPath)
}

func (c *FileCache) Put(sha string, sourcePath string) error {
	if !c.enabled || sha == "" {
		return nil
	}

	cachePath := c.cachePath(sha)
	if _, err := os.Stat(cachePath); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}

	if err := os.Link(sourcePath, cachePath); err == nil {
		return nil
	}

	_, err := c.copyFile(sourcePath, cachePath)
	return err
}

func (c *FileCache) cachePath(sha string) string {
	return filepath.Join(c.cacheDir, sha[:2], sha[2:4], sha)
}

func (c *FileCache) copyFile(src, dst string) (bool, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		os.Remove(dst)
		return false, err
	}

	return true, nil
}

func ComputeFileSHA(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

var globalCache *FileCache

func init() {
	globalCache, _ = NewFileCache()
}

func GetCache() *FileCache {
	return globalCache
}
