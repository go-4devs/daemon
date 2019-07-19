package daemon

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// Job interface for the do by manager
type Job interface {
	RunCtx(ctx context.Context) error
	CloseCtx(ctx context.Context) error
	fmt.Stringer
	Configure(opts ...Option) Setting
}

// JOption configure job
type JOption func(*job)

// NewJob create new job with options
func NewJob(r Run, opts ...JOption) Job {
	j := &job{
		r: r,
		c: func(ctx context.Context) error {
			return nil
		},
		cfg: newConfig(),
	}
	for _, opt := range opts {
		opt(j)
	}
	if j.name == "" {
		j.name = getFuncName(r)
	}
	if j.mw.close == nil {
		j.mw.close = DefaultHandle
	}
	if j.mw.run == nil {
		j.mw.run = DefaultHandle
	}

	return j
}

// WithJName set job name
func WithJName(name string) JOption {
	return func(i *job) {
		i.name = name
	}
}

// WithJClose set func which execute after job stop
func WithJClose(cl func(ctx context.Context) error) JOption {
	return func(i *job) {
		i.c = cl
	}
}

// WithJMiddleware add middleware to job
func WithJMiddleware(f ...Middleware) JOption {
	return func(i *job) {
		closes := make([]Handle, 0)
		runs := make([]Handle, 0)
		if i.mw.run != nil {
			runs = append(runs, i.mw.run)
		}
		if i.mw.close != nil {
			closes = append(closes, i.mw.close)
		}
		for _, p := range f {
			if p.close != nil {
				closes = append(closes, p.close)
			}
			if p.run != nil {
				runs = append(runs, p.run)
			}
		}
		if len(runs) > 0 {
			i.mw.run = chain(runs...)
		}
		if len(closes) > 0 {
			i.mw.close = chain(closes...)
		}
	}
}

// WithJOption set options job
func WithJOption(opts ...Option) JOption {
	return func(i *job) {
		for _, opt := range opts {
			opt(i.cfg)
		}
	}
}

func getFuncName(i interface{}) string {
	callerName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	if callerName == "" {
		return "-"
	}

	callerName = strings.NewReplacer("(*", "", ")", "").Replace(callerName)
	lastIndex := strings.LastIndex(callerName, "/")
	if lastIndex != -1 {
		callerName = callerName[lastIndex+1:]
	}

	return callerName
}

type job struct {
	r    func(ctx context.Context) error
	name string
	mw   Middleware
	cfg  *config
	c    func(ctx context.Context) error
}

func (j *job) RunCtx(ctx context.Context) error {
	return j.mw.run(ctx, j.r)
}

func (j *job) CloseCtx(ctx context.Context) error {
	j.cfg.timer.Stop()
	return j.mw.close(ctx, j.c)
}

func (j *job) String() string {
	return j.name
}

func (j *job) Configure(opts ...Option) Setting {
	for _, opt := range opts {
		opt(j.cfg)
	}
	return j.cfg
}
