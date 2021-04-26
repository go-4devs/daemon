package daemon

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// Option configure job.
type Option func(*Job)

// Job run job by frequency.
type Job struct {
	name      string
	run       Run
	stop      Run
	sem       chan struct{}
	err       chan error
	timer     Timer
	delay     time.Duration
	freq      func(time.Time) time.Duration
	handleErr func(error)
}

// WithName sets job name.
func WithName(name string) Option {
	return func(j *Job) { j.name = name }
}

// WithStop sets stop handle for job.
func WithStop(stop Run) Option {
	return func(j *Job) { j.stop = stop }
}

// WithTimer sets time,r to job.
func WithTimer(timer Timer) Option {
	return func(j *Job) { j.timer = timer }
}

// WithDelay sets delay Run job.
func WithDelay(delay time.Duration) Option {
	return func(j *Job) {
		j.timer.Reset(delay)
		j.delay = delay
	}
}

// WithFreq sets frequency Run job..
func WithFreq(freq time.Duration) Option {
	return func(j *Job) {
		j.freq = func(time.Time) time.Duration {
			return freq
		}
	}
}

// WithSchedule set delay and frequency Run job.
func WithSchedule(next func(time.Time) time.Duration) Option {
	return func(j *Job) {
		j.freq = next
		j.delay = next(time.Now())
		j.timer.Reset(j.delay)
	}
}

// WithRunMiddleware added middleware for the run job.
func WithRunMiddleware(fn ...Handle) Option {
	return func(j *Job) {
		run := j.run
		j.run = func(ctx context.Context) error {
			return chain(fn...)(ctx, run)
		}
	}
}

// WithStopMiddleware added middleware for the stop job.
func WithStopMiddleware(fn ...Handle) Option {
	return func(j *Job) {
		stop := j.stop
		j.stop = func(ctx context.Context) error {
			return chain(fn...)(ctx, stop)
		}
	}
}

// WithHandleErr add error hanler.
func WithHandleErr(fn func(error)) Option {
	return func(j *Job) {
		j.handleErr = fn
	}
}

// Run init function for the change state.
type Run func(ctx context.Context) error

// Handle middleware interface.
type Handle func(ctx context.Context, next Run) error

func stopJob(ctx context.Context) error {
	return nil
}

func handleErr(error) {}

// NewJob creates new job.
func NewJob(run Run, opts ...Option) *Job {
	j := Job{
		delay: time.Nanosecond,
		sem:   make(chan struct{}, 1),
		err:   make(chan error, 1),
		timer: NewTicker(time.Nanosecond),
		freq: func(time.Time) time.Duration {
			return time.Second
		},
		name:      getFuncName(run),
		stop:      stopJob,
		run:       run,
		handleErr: handleErr,
	}
	j = j.With(opts...)

	return &j
}

// HandleErr handle returned error.
func (j *Job) HandleErr(err error) {
	j.handleErr(err)
}

// Do run job.
func (j *Job) Do(ctx context.Context) <-chan error {
	<-j.timer.Tick()
	j.sem <- struct{}{}
	err := j.run(ctx)
	<-j.sem

	switch tr := err.(type) {
	case *stop:
	case *delay:
		j.timer.Reset(tr.d)
	default:
		j.timer.Reset(j.freq(time.Now()))
	}
	j.err <- err

	return j.err
}

// Stop job.
func (j *Job) Stop(ctx context.Context) error {
	return j.stop(ctx)
}

// String gets job name.
func (j *Job) String() string {
	return j.name
}

// With configure job.
func (j Job) With(opts ...Option) Job {
	for _, o := range opts {
		o(&j)
	}

	return j
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
