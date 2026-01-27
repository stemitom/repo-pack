package helpers

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DownloadStatus int

const (
	StatusPending DownloadStatus = iota
	StatusDownloading
	StatusCompleted
	StatusSkipped
	StatusFailed
)

type FileProgress struct {
	Path       string
	Size       int64
	Downloaded int64
	Status     DownloadStatus
	Error      error
}

type ProgressTracker struct {
	mu sync.Mutex

	startTime  time.Time
	totalFiles int64
	totalBytes int64

	files map[string]*FileProgress

	completedFiles int64
	completedBytes int64
	skippedFiles   int64
	failedFiles    int64

	width          int
	style          string
	maxActiveShow  int
	lastRender     time.Time
	lastLineCount  int
	renderInterval time.Duration
	quiet          bool
}

func NewProgressTracker(totalFiles int64, totalBytes int64) *ProgressTracker {
	return &ProgressTracker{
		startTime:      time.Now(),
		totalFiles:     totalFiles,
		totalBytes:     totalBytes,
		files:          make(map[string]*FileProgress),
		width:          40,
		style:          "#",
		maxActiveShow:  5,
		renderInterval: 100 * time.Millisecond,
	}
}

func (p *ProgressTracker) SetStyle(style string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if style != "" {
		p.style = style
	}
}

func (p *ProgressTracker) SetQuiet(quiet bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.quiet = quiet
}

func (p *ProgressTracker) SetMaxActiveDisplay(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxActiveShow = n
}

func (p *ProgressTracker) StartFile(path string, size int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.files[path] = &FileProgress{
		Path:   path,
		Size:   size,
		Status: StatusDownloading,
	}
	p.render()
}

func (p *ProgressTracker) UpdateFileProgress(path string, downloaded int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if fp, ok := p.files[path]; ok {
		fp.Downloaded = downloaded
	}
	p.render()
}

func (p *ProgressTracker) CompleteFile(path string, finalSize int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if fp, ok := p.files[path]; ok {
		fp.Status = StatusCompleted
		fp.Downloaded = finalSize
		if fp.Size == 0 {
			fp.Size = finalSize
		}
	}
	p.completedFiles++
	p.completedBytes += finalSize
	p.render()
}

func (p *ProgressTracker) SkipFile(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if fp, ok := p.files[path]; ok {
		fp.Status = StatusSkipped
	} else {
		p.files[path] = &FileProgress{
			Path:   path,
			Status: StatusSkipped,
		}
	}
	p.skippedFiles++
	p.render()
}

func (p *ProgressTracker) FailFile(path string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if fp, ok := p.files[path]; ok {
		fp.Status = StatusFailed
		fp.Error = err
	} else {
		p.files[path] = &FileProgress{
			Path:   path,
			Status: StatusFailed,
			Error:  err,
		}
	}
	p.failedFiles++
	p.render()
}

func (p *ProgressTracker) render() {
	if p.quiet {
		return
	}

	now := time.Now()
	if now.Sub(p.lastRender) < p.renderInterval {
		return
	}
	p.lastRender = now

	if p.lastLineCount > 0 {
		MoveCursorUp(p.lastLineCount)
	}

	var lines []string

	processedFiles := p.completedFiles + p.skippedFiles + p.failedFiles
	filePercent := 0
	if p.totalFiles > 0 {
		filePercent = int(float64(processedFiles) / float64(p.totalFiles) * 100)
	}

	barFilled := min(int(float64(processedFiles)/float64(p.totalFiles)*float64(p.width)), p.width)
	bar := strings.Repeat(p.style, barFilled) + strings.Repeat("-", p.width-barFilled)

	var bytesInfo string
	if p.totalBytes > 0 {
		bytesInfo = fmt.Sprintf(" %s/%s", FormatBytes(p.completedBytes), FormatBytes(p.totalBytes))
	}

	progressLine := fmt.Sprintf("Downloading [%s] %3d%% %s %d/%d files%s",
		Colorize(bar, Cyan),
		filePercent,
		Colorize("•", Dim),
		processedFiles,
		p.totalFiles,
		bytesInfo,
	)
	lines = append(lines, progressLine)

	elapsed := time.Since(p.startTime)
	filesPerSec := float64(p.completedFiles) / elapsed.Seconds()
	bytesPerSec := float64(p.completedBytes) / elapsed.Seconds()

	var eta string
	if filesPerSec > 0 && p.completedFiles > 0 {
		remainingFiles := p.totalFiles - processedFiles
		etaDuration := time.Duration(float64(remainingFiles)/filesPerSec) * time.Second
		eta = formatDuration(etaDuration)
	} else {
		eta = "calculating..."
	}

	statsLine := fmt.Sprintf("%s ETA %s %s %.1f files/s %s %s/s",
		Colorize("⏱", Dim),
		eta,
		Colorize("•", Dim),
		filesPerSec,
		Colorize("•", Dim),
		FormatBytes(int64(bytesPerSec)),
	)
	lines = append(lines, statsLine)

	statusParts := []string{}
	if p.completedFiles > 0 {
		statusParts = append(statusParts, Colorize(fmt.Sprintf("✓ %d completed", p.completedFiles), Green))
	}
	if p.skippedFiles > 0 {
		statusParts = append(statusParts, Colorize(fmt.Sprintf("⏭ %d skipped", p.skippedFiles), Yellow))
	}
	if p.failedFiles > 0 {
		statusParts = append(statusParts, Colorize(fmt.Sprintf("✗ %d failed", p.failedFiles), Red))
	}
	if len(statusParts) > 0 {
		lines = append(lines, strings.Join(statusParts, "  "))
	}

	activeFiles := p.getActiveFiles()
	if len(activeFiles) > 0 {
		lines = append(lines, "")
		for i, fp := range activeFiles {
			if i >= p.maxActiveShow {
				remaining := len(activeFiles) - p.maxActiveShow
				lines = append(lines, Colorize(fmt.Sprintf("  ... and %d more", remaining), Dim))
				break
			}
			filename := filepath.Base(fp.Path)
			if len(filename) > 30 {
				filename = filename[:27] + "..."
			}

			var progress string
			if fp.Size > 0 {
				progress = fmt.Sprintf("%s/%s", FormatBytes(fp.Downloaded), FormatBytes(fp.Size))
			} else if fp.Downloaded > 0 {
				progress = FormatBytes(fp.Downloaded)
			} else {
				progress = "starting..."
			}
			lines = append(lines, fmt.Sprintf("%s %s (%s)",
				Colorize("↓", Cyan),
				filename,
				Colorize(progress, Dim),
			))
		}
	}

	for _, line := range lines {
		ClearLine()
		fmt.Println(line)
	}

	p.lastLineCount = len(lines)
}

func (p *ProgressTracker) getActiveFiles() []*FileProgress {
	var active []*FileProgress
	for _, fp := range p.files {
		if fp.Status == StatusDownloading {
			active = append(active, fp)
		}
	}
	return active
}

func (p *ProgressTracker) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet {
		return
	}

	maxPossibleLines := 4 + p.maxActiveShow + 1
	linesToClear := max(p.lastLineCount, maxPossibleLines)

	MoveCursorUp(linesToClear)
	ClearScreenBelow()

	elapsed := time.Since(p.startTime)

	var summaryLines []string
	summaryLines = append(summaryLines, Colorize("✓ Download complete!", Green+Bold))
	summaryLines = append(summaryLines, "")
	summaryLines = append(summaryLines, fmt.Sprintf("  %d/%d files %s %s %s %s",
		p.completedFiles,
		p.totalFiles,
		Colorize("•", Dim),
		FormatBytes(p.completedBytes),
		Colorize("•", Dim),
		elapsed.Round(time.Millisecond),
	))

	statusParts := []string{}
	if p.completedFiles > 0 {
		statusParts = append(statusParts, Colorize(fmt.Sprintf("✓ %d downloaded", p.completedFiles), Green))
	}
	if p.skippedFiles > 0 {
		statusParts = append(statusParts, Colorize(fmt.Sprintf("⏭ %d skipped", p.skippedFiles), Yellow))
	}
	if p.failedFiles > 0 {
		statusParts = append(statusParts, Colorize(fmt.Sprintf("✗ %d failed", p.failedFiles), Red))
	}
	if len(statusParts) > 0 {
		summaryLines = append(summaryLines, "  "+strings.Join(statusParts, "  "))
	}
	summaryLines = append(summaryLines, "")

	for _, line := range summaryLines {
		fmt.Println(line)
	}
}

func (p *ProgressTracker) GetStats() (completed, skipped, failed int64, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.completedFiles, p.skippedFiles, p.failedFiles, time.Since(p.startTime)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("~%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		if secs > 0 {
			return fmt.Sprintf("%dm %ds", mins, secs)
		}
		return fmt.Sprintf("%dm", mins)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, mins)
}

type SingleFileProgress struct {
	mu         sync.Mutex
	filename   string
	totalBytes int64
	downloaded int64
	startTime  time.Time
	lastRender time.Time
	quiet      bool
}

func NewSingleFileProgress(filename string, totalBytes int64) *SingleFileProgress {
	return &SingleFileProgress{
		filename:   filename,
		totalBytes: totalBytes,
		startTime:  time.Now(),
	}
}

func (s *SingleFileProgress) SetQuiet(quiet bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quiet = quiet
}

func (s *SingleFileProgress) Update(downloaded int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.quiet {
		return
	}

	now := time.Now()
	if now.Sub(s.lastRender) < 100*time.Millisecond {
		return
	}
	s.lastRender = now
	s.downloaded = downloaded

	s.render()
}

func (s *SingleFileProgress) render() {
	width := 30
	percent := 0
	if s.totalBytes > 0 {
		percent = int(float64(s.downloaded) / float64(s.totalBytes) * 100)
	}

	barFilled := 0
	if s.totalBytes > 0 {
		barFilled = int(float64(s.downloaded) / float64(s.totalBytes) * float64(width))
	}
	if barFilled > width {
		barFilled = width
	}
	bar := strings.Repeat("#", barFilled) + strings.Repeat("-", width-barFilled)

	elapsed := time.Since(s.startTime)
	bytesPerSec := float64(s.downloaded) / elapsed.Seconds()

	var eta string
	if s.totalBytes > 0 && bytesPerSec > 0 {
		remaining := s.totalBytes - s.downloaded
		etaDuration := time.Duration(float64(remaining)/bytesPerSec) * time.Second
		eta = formatDuration(etaDuration)
	} else {
		eta = "..."
	}

	var sizeInfo string
	if s.totalBytes > 0 {
		sizeInfo = fmt.Sprintf("%s/%s", FormatBytes(s.downloaded), FormatBytes(s.totalBytes))
	} else {
		sizeInfo = FormatBytes(s.downloaded)
	}

	ClearLine()
	fmt.Printf("%s Downloading [%s] %3d%% %s %s %s/s %s ETA %s",
		Colorize("↓", Cyan),
		Colorize(bar, Cyan),
		percent,
		Colorize("•", Dim),
		sizeInfo,
		Colorize(FormatBytes(int64(bytesPerSec)), Dim),
		Colorize("•", Dim),
		eta,
	)
}

func (s *SingleFileProgress) Finish() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.quiet {
		return
	}

	elapsed := time.Since(s.startTime)

	ClearLine()
	fmt.Println()
	fmt.Printf("%s Downloaded %s (%s) in %s\n",
		Colorize("✓", Green),
		s.filename,
		FormatBytes(s.downloaded),
		elapsed.Round(time.Millisecond),
	)
}

type Bar struct {
	tracker *ProgressTracker
}

func (bar *Bar) Config(start, total int64, description string) {
	bar.tracker = NewProgressTracker(total, 0)
}

func (bar *Bar) SetStyle(style string) {
	if bar.tracker != nil {
		bar.tracker.SetStyle(style)
	}
}

func (bar *Bar) Update(cur int64) {
}

func (bar *Bar) Increment() {
	if bar.tracker != nil {
		bar.tracker.mu.Lock()
		bar.tracker.completedFiles++
		bar.tracker.render()
		bar.tracker.mu.Unlock()
	}
}

func (bar *Bar) Finish() {
	if bar.tracker != nil {
		bar.tracker.Finish()
	}
}
