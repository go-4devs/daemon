package daemon

import (
	"context"
	"io"
	"sync"
)

var defaultHandle Handle = func(ctx context.Context, next Run) error {
	return next(ctx)
}

// Run init function for the change state
type Run func(ctx context.Context) error

// Handle middleware interface
type Handle func(ctx context.Context, next Run) error

//Manager monitoring do job
type Manager interface {
	DoCtx(ctx context.Context, j Job, o ...Option)
	Wait()
	io.Closer
}

//MOption configure manager
type MOption func(*manager)

//NewManager create new manager with options
func NewManager(opts ...MOption) Manager {
	m := &manager{
		close:    make(chan struct{}),
		closeErr: func(err error) {},
	}
	for _, opt := range opts {
		opt(m)
	}
	if m.closeHandle == nil {
		m.closeHandle = defaultHandle
	}
	if m.runHandle == nil {
		m.runHandle = defaultHandle
	}

	return m
}

//WithMMiddleware set middleware to manager
func WithMMiddleware(m ...Middleware) MOption {
	return func(i *manager) {
		runs := make([]Handle, 0)
		closes := make([]Handle, 0)
		if i.runHandle != nil {
			runs = append(runs, i.runHandle)
		}
		if i.closeHandle != nil {
			closes = append(closes, i.closeHandle)
		}
		for _, r := range m {
			if r.close != nil {
				closes = append(closes, r.close)
			}
			if r.run != nil {
				runs = append(runs, r.run)
			}
		}
		if len(runs) > 0 {
			i.runHandle = chain(runs...)
		}
		if len(closes) > 0 {
			i.closeHandle = chain(closes...)
		}
	}
}

//WithHandleCloseErr Handle close err
func WithHandleCloseErr(f func(err error)) MOption {
	return func(i *manager) {
		i.closeErr = f
	}
}

type manager struct {
	wg          sync.WaitGroup
	close       chan struct{}
	closeHandle Handle
	runHandle   Handle
	closeErr    func(err error)
}

func (m *manager) DoCtx(ctx context.Context, j Job, opts ...Option) {
	m.wg.Add(1)
	s := j.Configure(opts...)
	go func() {
		defer func() {
			m.closeErr(m.closeHandle(ctx, j.CloseCtx))
			m.wg.Done()
		}()
		for {
			select {
			case <-m.close:
				return
			case <-ctx.Done():
				return
			case _, ok := <-s.Do():
				if ok && !s.IsProcessed(m.runHandle(ctx, j.RunCtx)) {
					return
				}
			}
		}
	}()
}

func (m *manager) Wait() {
	m.wg.Wait()
}

func (m *manager) Close() error {
	close(m.close)
	m.Wait()
	return nil
}

func chain(handleFunc ...Handle) Handle {
	n := len(handleFunc)

	if n > 1 {
		lastI := n - 1
		return func(ctx context.Context, next Run) error {
			var (
				chainHandler Run
				curI         int
			)
			chainHandler = func(currentCtx context.Context) error {
				if curI == lastI {
					return next(currentCtx)
				}
				curI++
				err := handleFunc[curI](currentCtx, chainHandler)
				curI--
				return err

			}
			return handleFunc[0](ctx, chainHandler)
		}
	}

	if n == 1 {
		return handleFunc[0]
	}

	return func(ctx context.Context, next Run) error {
		return next(ctx)
	}
}
