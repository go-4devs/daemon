package daemon

import "time"

// StopJob stop job
func StopJob(err error) error {
	return &stop{e: err}
}

// IsStoppedJob check stopped job
func IsStoppedJob(err error) bool {
	_, ok := err.(*stop)
	return ok
}

type stop struct {
	e error
}

func (s *stop) Error() string {
	return s.e.Error()
}

// GetDelayedJob get delay job
func GetDelayedJob(err error) (time.Duration, bool) {
	if d, ok := err.(*delay); ok {
		return d.d, ok
	}
	return 0, false
}

// DelayJob update delay next Run job
func DelayJob(d time.Duration, err error) error {
	return &delay{d: d, e: err}
}

type delay struct {
	d time.Duration
	e error
}

func (d *delay) Error() string {
	return d.e.Error()
}
