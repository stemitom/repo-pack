package helpers

import "fmt"

type Bar struct {
	rate    string
	graph   string
	percent int64
	Cur     int64
	total   int64
}

func (bar *Bar) Config(start, total int64) {
	bar.Cur = start
	bar.total = total
	if bar.graph == "" {
		bar.graph = "*"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph
	}
}

func (bar *Bar) getPercent() int64 {
	return int64((float32(bar.Cur) / float32(bar.total)) * 100)
}

func (bar *Bar) Update(cur int64) {
	bar.Play(cur)
}

func (bar *Bar) Play(cur int64) {
	bar.Cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
	fmt.Printf("\r[%-50s]%3d%% %8d/%d", bar.rate, bar.percent, bar.Cur, bar.total)
}

func (bar *Bar) Finish() {
	fmt.Println()
}
