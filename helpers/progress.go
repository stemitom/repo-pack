package helpers

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Bar struct {
	mu          sync.Mutex
	startTime   time.Time
	rate        string
	graph       string
	description string
	percent     int64
	cur         int64
	total       int64
	width       int
	lastUpdate  time.Time
}

func (bar *Bar) Config(start, total int64, description string) {
	bar.mu.Lock()
	defer bar.mu.Unlock()
	bar.cur = start
	bar.total = total
	bar.width = 50
	bar.graph = "█"
	bar.description = description
	bar.startTime = time.Now()
	bar.updateRate()
}

func (bar *Bar) SetStyle(style string) {
	bar.mu.Lock()
	defer bar.mu.Unlock()
	if style != "" {
		bar.graph = style
	}
}

func (bar *Bar) getPercent() int64 {
	return int64((float64(bar.cur) / float64(bar.total)) * 100)
}

func (bar *Bar) updateRate() {
	completedWidth := int((float64(bar.cur) / float64(bar.total)) * float64(bar.width))
	bar.rate = strings.Repeat(bar.graph, completedWidth) + strings.Repeat(" ", bar.width-completedWidth)
}

func (bar *Bar) Update(cur int64) {
	bar.mu.Lock()
	defer bar.mu.Unlock()
	bar.cur = cur
	bar.play()
}

// Increment atomically increments the progress counter by 1
func (bar *Bar) Increment() {
	bar.mu.Lock()
	defer bar.mu.Unlock()
	bar.cur++
	bar.play()
}

func (bar *Bar) play() {
	lastPercent := bar.percent
	bar.percent = bar.getPercent()

	// Reduce flickering
	now := time.Now()
	if now.Sub(bar.lastUpdate) < 100*time.Millisecond {
		return
	}
	bar.lastUpdate = now

	if bar.percent != lastPercent {
		bar.updateRate()
	}
	elapsedTime := time.Since(bar.startTime)
	itemsPerSec := float64(bar.cur) / elapsedTime.Seconds()
	fmt.Printf("\r%s |%-50s| %3d%% %3d/%d  %.2f it/s", bar.description, bar.rate, bar.percent, bar.cur, bar.total, itemsPerSec)
}

func (bar *Bar) Finish() {
	bar.mu.Lock()
	defer bar.mu.Unlock()
	bar.updateRate()
	elapsedTime := time.Since(bar.startTime)
	fmt.Printf("\r%s |%-50s| 100%% %3d/%d  Time: %s\n", bar.description, bar.rate, bar.total, bar.total, elapsedTime.String())
}
