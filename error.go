package daemon

import "time"

//StopJob stop job
func StopJob(err error) error {
	return &stop{e: err}
}

type stop struct {
	e error
}

func (s *stop) Error() string {
	return s.e.Error()
}

//DelayJob update delay next Run job
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
