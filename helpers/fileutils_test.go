package helpers

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkExtractRelativePath(b *testing.B) {
	baseDir := "src"
	filePath := "project/src/components/Button/index.tsx"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractRelativePath(baseDir, filePath)
	}
}

func BenchmarkFileExists(b *testing.B) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bench-fileutils-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "src", "test.txt")
	os.MkdirAll(filepath.Dir(testFile), 0o755)
	os.WriteFile(testFile, []byte("test"), 0o644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FileExists("src", "src/test.txt", tmpDir)
	}
}

func BenchmarkSaveFile(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-savefile-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := bytes.Repeat([]byte("test content"), 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := io.NopCloser(bytes.NewReader(content))
		_ = SaveFile("src", "src/file.txt", reader, tmpDir)
	}
}

func BenchmarkFilepathAbs(b *testing.B) {
	path := "/home/user/project/output"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = filepath.Abs(path)
	}
}

func BenchmarkFilepathAbsRelative(b *testing.B) {
	path := "./output/files"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = filepath.Abs(path)
	}
}
