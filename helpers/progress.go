package helpers

import (
	"fmt"
	"strings"
	"time"
)

type Bar struct {
	startTime   time.Time
	rate        string
	graph       string
	description string
	percent     int64
	Cur         int64
	total       int64
	width       int
}

func (bar *Bar) Config(start, total int64, description string) {
	bar.Cur = start
	bar.total = total
	bar.width = 50
	bar.graph = "█"
	bar.description = description
	bar.startTime = time.Now()
	bar.updateRate()
}

func (bar *Bar) getPercent() int64 {
	return int64((float64(bar.Cur) / float64(bar.total)) * 100)
}

func (bar *Bar) updateRate() {
	completedWidth := int((float64(bar.Cur) / float64(bar.total)) * float64(bar.width))
	bar.rate = strings.Repeat(bar.graph, completedWidth) + strings.Repeat(" ", bar.width-completedWidth)
}

func (bar *Bar) Update(cur int64) {
	bar.Cur = cur
	bar.Play(cur)
}

func (bar *Bar) Play(cur int64) {
	bar.Cur = cur
	lastPercent := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != lastPercent {
		bar.updateRate()
	}
	elapsedTime := time.Since(bar.startTime)
	itemsPerSec := float64(bar.Cur) / elapsedTime.Seconds()
	fmt.Printf("\r%s |%-50s| %3d%% %3d/%d %.2f it/s", bar.description, bar.rate, bar.percent, bar.Cur, bar.total, itemsPerSec)
}

func (bar *Bar) Finish() {
	bar.updateRate()
	elapsedTime := time.Since(bar.startTime)
	fmt.Printf("\r%s |%-50s| 100%% %3d/%d  Time: %s\n", bar.description, bar.rate, bar.total, bar.total, elapsedTime.String())
}
