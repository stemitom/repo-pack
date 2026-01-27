package helpers_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"repo-pack/helpers"
)

func TestProgressTrackerBasic(t *testing.T) {
	helpers.SetColorEnabled(false)

	tracker := helpers.NewProgressTracker(10, 1024*1024)
	tracker.SetQuiet(true)

	tracker.StartFile("file1.txt", 1024)
	tracker.UpdateFileProgress("file1.txt", 512)
	tracker.CompleteFile("file1.txt", 1024)

	completed, skipped, failed, _ := tracker.GetStats()
	assert.Equal(t, int64(1), completed)
	assert.Equal(t, int64(0), skipped)
	assert.Equal(t, int64(0), failed)
}

func TestProgressTrackerSkipAndFail(t *testing.T) {
	helpers.SetColorEnabled(false)

	tracker := helpers.NewProgressTracker(5, 0)
	tracker.SetQuiet(true)

	tracker.CompleteFile("file1.txt", 100)
	tracker.SkipFile("file2.txt")
	tracker.FailFile("file3.txt", nil)
	tracker.CompleteFile("file4.txt", 200)
	tracker.SkipFile("file5.txt")

	completed, skipped, failed, _ := tracker.GetStats()
	assert.Equal(t, int64(2), completed)
	assert.Equal(t, int64(2), skipped)
	assert.Equal(t, int64(1), failed)
}

func TestProgressTrackerConcurrent(t *testing.T) {
	helpers.SetColorEnabled(false)

	tracker := helpers.NewProgressTracker(100, 0)
	tracker.SetQuiet(true)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				path := "file.txt"
				tracker.StartFile(path, 1000)
				tracker.UpdateFileProgress(path, 500)
				tracker.CompleteFile(path, 1000)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	completed, _, _, _ := tracker.GetStats()
	assert.Equal(t, int64(100), completed)
}

func TestProgressTrackerSetStyle(t *testing.T) {
	tracker := helpers.NewProgressTracker(10, 0)
	tracker.SetStyle("*")
	tracker.SetQuiet(true)
	tracker.CompleteFile("test.txt", 100)
}

func TestProgressTrackerSetMaxActiveDisplay(t *testing.T) {
	tracker := helpers.NewProgressTracker(10, 0)
	tracker.SetMaxActiveDisplay(3)
	tracker.SetQuiet(true)
	tracker.CompleteFile("test.txt", 100)
}

func TestSingleFileProgressBasic(t *testing.T) {
	helpers.SetColorEnabled(false)

	progress := helpers.NewSingleFileProgress("test.txt", 1024)
	progress.SetQuiet(true)

	progress.Update(256)
	progress.Update(512)
	progress.Update(1024)
	progress.Finish()
}

func TestSingleFileProgressUnknownSize(t *testing.T) {
	helpers.SetColorEnabled(false)

	progress := helpers.NewSingleFileProgress("test.txt", 0)
	progress.SetQuiet(true)

	progress.Update(100)
	progress.Update(500)
	progress.Finish()
}

func TestLegacyBarCompatibility(t *testing.T) {
	helpers.SetColorEnabled(false)

	bar := &helpers.Bar{}
	bar.Config(0, 10, "Test: ")
	bar.SetStyle("*")

	for i := 0; i < 10; i++ {
		bar.Increment()
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := helpers.FormatBytes(tt.bytes)
		assert.Equal(t, tt.expected, result, "FormatBytes(%d)", tt.bytes)
	}
}

func TestBarFinishOutput(t *testing.T) {
	helpers.SetColorEnabled(false)

	tracker := helpers.NewProgressTracker(5, 5000)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.CompleteFile("file1.txt", 1000)
	tracker.CompleteFile("file2.txt", 1000)
	tracker.SkipFile("file3.txt")
	tracker.FailFile("file4.txt", nil)
	tracker.CompleteFile("file5.txt", 1000)
	tracker.Finish()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Download complete")
	assert.Contains(t, output, "3 downloaded")
	assert.Contains(t, output, "1 skipped")
	assert.Contains(t, output, "1 failed")
}

func TestColorSupport(t *testing.T) {
	originalEnv := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalEnv)

	helpers.SetColorEnabled(true)
	assert.True(t, helpers.SupportsColor())

	helpers.SetColorEnabled(false)
	assert.False(t, helpers.SupportsColor())
}

func TestColorize(t *testing.T) {
	helpers.SetColorEnabled(true)
	colored := helpers.Colorize("test", helpers.Green)
	assert.Contains(t, colored, "test")
	assert.Contains(t, colored, "\033[")

	helpers.SetColorEnabled(false)
	plain := helpers.Colorize("test", helpers.Green)
	assert.Equal(t, "test", plain)
}
