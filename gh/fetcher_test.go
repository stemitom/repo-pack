package gh

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

// mockResponse creates a mock HTTP response for testing
func mockResponse(body string, contentLength int) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Length": []string{string(rune(contentLength))},
		},
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}
}

// mockLfsPointer returns a typical LFS pointer content
func mockLfsPointer() string {
	return "version https://git-lfs.github.com/spec/v1\noid sha256:abc123\nsize 1234567"
}

// mockRegularContent returns non-LFS content
func mockRegularContent(size int) string {
	return string(make([]byte, size))
}

func BenchmarkIsLfsResponse_LfsPointer(b *testing.B) {
	lfsContent := mockLfsPointer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Length": []string{"134"},
			},
			Body: io.NopCloser(bytes.NewBufferString(lfsContent)),
		}
		_ = isLfsResponse(resp)
	}
}

func BenchmarkIsLfsResponse_SmallFile(b *testing.B) {
	content := mockRegularContent(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Length": []string{"500"},
			},
			Body: io.NopCloser(bytes.NewBufferString(content)),
		}
		_ = isLfsResponse(resp)
	}
}

func BenchmarkIsLfsResponse_LargeFile(b *testing.B) {
	// Simulate checking a 1MB file header
	content := mockRegularContent(1024 * 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Length": []string{"1048576"},
			},
			Body: io.NopCloser(bytes.NewBufferString(content)),
		}
		_ = isLfsResponse(resp)
	}
}
