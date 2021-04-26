package daemon

import (
	"context"
	"errors"
	"sync"
)

// Manager run jobs.
type Manager struct {
	sync.WaitGroup
	close chan struct{}
	opts  []Option
}

// New creates new manager and configure them.
func New(opts ...Option) *Manager {
	m := &Manager{
		opts:  opts,
		close: make(chan struct{}),
	}

	return m
}

// Do runs job.
func (m *Manager) Do(ctx context.Context, j *Job, opts ...Option) {
	m.Add(1)
	job := j.With(append(m.opts, opts...)...)

	go func() {
		defer func() {
			j.HandleErr(job.Stop(ctx))
			m.Done()
		}()

		for {
			select {
			case <-m.close:
				return
			case <-ctx.Done():
				return
			case err := <-job.Do(ctx):
				if err != nil {
					if errors.Is(err, &stop{}) {
						return
					}

					j.HandleErr(err)
				}
			}
		}
	}()
}

// Close jobs.
func (m *Manager) Close() error {
	close(m.close)
	m.Wait()

	return nil
}
