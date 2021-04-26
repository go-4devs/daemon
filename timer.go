package daemon

import "time"

// Timer for the Run job.
type Timer interface {
	Tick() <-chan time.Time
	Reset(d time.Duration)
	Stop()
}

// NewTicker create new ticker based on time.ticker.
func NewTicker(freq time.Duration) Timer {
	return &ticker{
		freq:   freq,
		ticker: time.NewTicker(freq),
	}
}

type ticker struct {
	freq   time.Duration
	ticker *time.Ticker
}

// Tick time.
func (t *ticker) Tick() <-chan time.Time {
	return t.ticker.C
}

// Stop timer.
func (t *ticker) Stop() {
	t.ticker.Stop()
}

// Reset timer.
func (t *ticker) Reset(freq time.Duration) {
	if t.freq != freq && freq > 0 {
		t.ticker.Stop()
		t.freq = freq
		t.ticker = time.NewTicker(freq)
	}
}
