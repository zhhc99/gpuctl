package stress

import (
	"fmt"
	"sync"
	"time"
)

type Display struct {
	stats  *Stats
	ticker *time.Ticker
	done   chan struct{}
	wg     sync.WaitGroup
}

func NewDisplay(stats *Stats) *Display {
	return &Display{
		stats: stats,
		done:  make(chan struct{}),
	}
}

func (d *Display) Start() {
	s := d.stats
	fmt.Printf("Stressing %s | Mode: %s | Ctrl+C to stop\n", s.Name, s.Mode)
	fmt.Printf("  %-16s %-12s %s\n", "Dispatches/s", "Buffer", "Elapsed")
	d.printRow()

	d.ticker = time.NewTicker(time.Second)
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-d.ticker.C:
				d.refresh()
			case <-d.done:
				return
			}
		}
	}()
}

func (d *Display) refresh() {
	fmt.Printf("\033[1A\r\033[K")
	d.printRow()
}

func (d *Display) printRow() {
	s := d.stats
	fmt.Printf("  %-16s %-12s %s\n",
		fmtInt(s.Rate()),
		fmtBytes(s.Buffer),
		fmtDuration(s.Elapsed()),
	)
}

func (d *Display) Stop() {
	d.ticker.Stop()
	close(d.done)
	d.wg.Wait()
	d.refresh()

	s := d.stats
	fmt.Printf("\nDone. %s elapsed | %s total dispatches | avg %s dispatches/s\n",
		fmtDuration(s.Elapsed()),
		fmtInt(s.Total()),
		fmtInt(s.AvgRate()),
	)
}

func fmtInt(n int64) string {
	if n < 0 {
		n = 0
	}
	if n < 1_000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%d,%03d", n/1_000, n%1_000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1_000_000, (n/1_000)%1_000, n%1_000)
}

func fmtBytes(b uint64) string {
	if b >= 1<<30 {
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	}
	return fmt.Sprintf("%.0f MB", float64(b)/float64(1<<20))
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
