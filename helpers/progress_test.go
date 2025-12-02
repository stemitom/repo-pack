package helpers_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"repo-pack/helpers"
	"github.com/stretchr/testify/assert"
)

func TestBarConfig(t *testing.T) {
	bar := &helpers.Bar{}
	bar.Config(0, 100, "Test: ")

	// Verify internal state after config
	// We can't directly access private fields, but we can test behavior
	bar.Increment()
	// If no panic, test passes
}

func TestBarIncrement(t *testing.T) {
	bar := &helpers.Bar{}
	bar.Config(0, 10, "Test: ")

	for i := 0; i < 10; i++ {
		bar.Increment()
	}
	bar.Finish()
	// If no panic, test passes
}

func TestBarSetStyle(t *testing.T) {
	bar := &helpers.Bar{}
	bar.Config(0, 100, "Test: ")
	bar.SetStyle("*")
	bar.Increment()
	// If no panic, test passes
}

func TestBarEmptyStyle(t *testing.T) {
	bar := &helpers.Bar{}
	bar.Config(0, 100, "Test: ")
	bar.SetStyle("")
	bar.Increment()
	// Should use default style, not empty
}

func TestBarUpdate(t *testing.T) {
	bar := &helpers.Bar{}
	bar.Config(0, 100, "Test: ")
	bar.Update(50)
	bar.Finish()
	// If no panic, test passes
}

func TestBarFinish(t *testing.T) {
	bar := &helpers.Bar{}
	bar.Config(0, 5, "Test: ")

	for i := 0; i < 5; i++ {
		bar.Increment()
	}

	// Capture output to verify it doesn't error
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	bar.Finish()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "100%")
	assert.NotEmpty(t, output)
}

func TestBarConcurrentIncrements(t *testing.T) {
	bar := &helpers.Bar{}
	bar.Config(0, 100, "Test: ")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				bar.Increment()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	bar.Finish()
	// If no panic with concurrent access, test passes
}
