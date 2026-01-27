package helpers

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	White  = "\033[37m"

	ClearLineSeq        = "\033[K"
	ClearScreenBelowSeq = "\033[J"
	CursorUpSeq         = "\033[%dA"
	CursorDownSeq       = "\033[%dB"
)

var colorEnabled = detectColorSupport()

func detectColorSupport() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	if runtime.GOOS == "windows" {
		return os.Getenv("WT_SESSION") != "" ||
			os.Getenv("TERM_PROGRAM") == "vscode" ||
			os.Getenv("ANSICON") != ""
	}
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func SupportsColor() bool {
	return colorEnabled
}

func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

func Colorize(text, color string) string {
	if !colorEnabled {
		return text
	}
	return color + text + Reset
}

func ClearLine() {
	if colorEnabled {
		fmt.Print("\r" + ClearLineSeq)
	} else {
		fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
	}
}

func MoveCursorUp(n int) {
	if colorEnabled && n > 0 {
		fmt.Printf(CursorUpSeq, n)
	}
}

func MoveCursorDown(n int) {
	if colorEnabled && n > 0 {
		fmt.Printf(CursorDownSeq, n)
	}
}

func ClearLines(n int) {
	for i := 0; i < n; i++ {
		ClearLine()
		if i < n-1 {
			MoveCursorUp(1)
		}
	}
}

func ClearScreenBelow() {
	if colorEnabled {
		fmt.Print(ClearScreenBelowSeq)
	}
}

func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
