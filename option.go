package daemon

import "time"

// Option for the do job
type Option func(*config)

// Setting job settings
type Setting interface {
	IsProcessed(err error) bool
	Do() <-chan time.Time
	Reload(s func(time.Time) time.Duration)
}

// WithSchedule set delay and frequency Run job
func WithSchedule(next func(time.Time) time.Duration) Option {
	return func(config *config) {
		config.Reload(next)
	}
}

// WithDelay set delay Run job
func WithDelay(delay time.Duration) Option {
	return func(config *config) {
		config.timer.Reset(delay)
		config.delay = delay
	}
}

// WithFreq set Frequency Run job
func WithFreq(freq time.Duration) Option {
	return func(config *config) {
		config.freq = func(time.Time) time.Duration {
			return freq
		}
	}
}

// WithHandleErr replace Handle errors
func WithHandleErr(f func(err error)) Option {
	return func(config *config) {
		config.hErr = f
	}
}

func newConfig() *config {
	return &config{
		delay: time.Nanosecond,
		hErr:  func(err error) {},
		timer: NewTicker(time.Nanosecond),
		freq: func(time.Time) time.Duration {
			return time.Second
		},
	}
}

type config struct {
	delay time.Duration
	freq  func(time.Time) time.Duration
	hErr  func(err error)
	timer Timer
}

func (o *config) Reload(s func(time.Time) time.Duration) {
	o.freq = s
	o.delay = s(time.Now())
	o.timer.Reset(o.delay)
}

func (o *config) IsProcessed(err error) bool {
	if err != nil {
		o.hErr(err)
	}
	switch tr := err.(type) {
	case *stop:
		return false
	case *delay:
		o.timer.Reset(tr.d)
	default:
		o.timer.Reset(o.freq(time.Now()))
	}
	return true
}

func (o *config) Do() <-chan time.Time {
	return o.timer.Tick()
}
