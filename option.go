package daemon

import "time"

//Option for the do job
type Option func(*config)

//Setting job settings
type Setting interface {
	IsProcessed(err error) bool
	Do() <-chan struct{}
}

//WithReload set reload freq
func WithReload(r func() <-chan time.Duration) Option {
	return func(config *config) {
		config.reload = r
	}
}

//WithDelay set delay Run job
func WithDelay(delay time.Duration) Option {
	return func(config *config) {
		config.timer.Reset(delay)
		config.delay = delay
	}
}

//WithFreq set Frequency Run job
func WithFreq(freq time.Duration) Option {
	return func(config *config) {
		config.freq = freq
	}
}

//WithHandleErr replace Handle errors
func WithHandleErr(f func(err error)) Option {
	return func(config *config) {
		config.hErr = f
	}
}

func newConfig() *config {
	return &config{
		reload: func() <-chan time.Duration {
			return nil
		},
		close: make(chan struct{}),
		do:    make(chan struct{}),
		delay: time.Nanosecond,
		hErr:  func(err error) {},
		timer: NewTicker(time.Nanosecond),
		freq:  time.Second,
	}
}

type config struct {
	delay  time.Duration
	freq   time.Duration
	reload func() <-chan time.Duration
	hErr   func(err error)
	timer  Timer
	close  chan struct{}
	do     chan struct{}
}

func (o *config) IsProcessed(err error) bool {
	o.hErr(err)
	switch tr := err.(type) {
	case *stop:
		return false
	case *delay:
		o.timer.Reset(tr.d)
	default:
		o.timer.Reset(o.freq)
	}
	return true
}

func (o *config) Do() <-chan struct{} {
	go func() {
		for {
			select {
			case t := <-o.reload():
				o.timer.Reset(t)
			case _, ok := <-o.timer.Tick():
				if ok {
					o.do <- struct{}{}
				}
			}
		}
	}()
	return o.do
}
