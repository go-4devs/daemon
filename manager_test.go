package daemon_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"gitoa.ru/go-4devs/daemon"
)

//nolint: gochecknoglobals
var (
	ctx           = context.Background()
	errJob        = errors.New("error job")
	jobName       = "job name"
	jobNameOption = daemon.WithName(jobName)
)

func createJob(count *int32, d time.Duration, err error, opts ...daemon.Option) *daemon.Job {
	opts = append(opts, jobNameOption)

	return daemon.NewJob(func(ctx context.Context) error {
		if d > 0 {
			<-time.After(d)
		}
		atomic.AddInt32(count, 1)
		return err
	}, opts...)
}

func TestDoJobSuccess(t *testing.T) {
	t.Parallel()

	var (
		cnt  int32
		cnt2 int32
		runm int32
		clm  int32
	)

	m := daemon.New()

	m.Do(ctx, createJob(&cnt, 0, nil), daemon.WithFreq(time.Second/3))

	mwRun := func(ctx context.Context, next daemon.Run) error {
		atomic.AddInt32(&runm, 1)
		return next(ctx)
	}
	mwStop := func(ctx context.Context, next daemon.Run) error {
		atomic.AddInt32(&clm, 1)
		return next(ctx)
	}
	m.Do(ctx, createJob(&cnt2, 0, nil,
		daemon.WithRunMiddleware(mwRun, mwRun, mwRun),
		daemon.WithStopMiddleware(mwStop, mwStop, mwStop),
	), daemon.WithFreq(time.Second/10))
	time.AfterFunc(time.Second, func() {
		requireNil(t, m.Close())
	})
	m.Wait()

	requireTrue(t, cnt > 2)
	requireTrue(t, cnt2 >= 10)
	requireTrue(t, cnt2*3 == runm)
	requireTrue(t, clm == 3)
}

func requireNil(t *testing.T, ex interface{}) {
	t.Helper()

	if ex != nil {
		t.Fatal("expect nil")
	}
}

func requireTrue(t *testing.T, ex bool) {
	t.Helper()

	if !ex {
		t.Fatal("expect true")
	}
}

func TestDoJobStop(t *testing.T) {
	t.Parallel()

	m := daemon.New()

	var cnt int32

	m.Do(ctx, createJob(&cnt, 0, daemon.StopJob(nil)), daemon.WithFreq(time.Nanosecond))
	m.Wait()
	requireTrue(t, cnt == 1)
}

func TestDoJobDelay(t *testing.T) {
	t.Parallel()

	var cnt int32

	m := daemon.New()

	m.Do(ctx, createJob(&cnt, 0, daemon.DelayJob(time.Second/3, nil)), daemon.WithFreq(time.Second))
	time.AfterFunc(time.Second, func() {
		requireNil(t, m.Close())
	})
	m.Wait()
	requireTrue(t, cnt > 2)
}

func TestDoJobSkipErr(t *testing.T) {
	t.Parallel()

	var cnt int32

	m := daemon.New()

	m.Do(ctx, createJob(&cnt, 0, errJob), daemon.WithFreq(time.Second/3))
	time.AfterFunc(time.Second, func() {
		requireNil(t, m.Close())
	})
	m.Wait()
	requireTrue(t, cnt > 2)
}

func TestDoJobRetryErr(t *testing.T) {
	t.Parallel()

	var cnt int32

	m := daemon.New()

	m.Do(ctx, createJob(&cnt, time.Millisecond, errJob, daemon.Retry(3, daemon.StopJob)), daemon.WithFreq(time.Nanosecond))
	m.Wait()
	requireTrue(t, cnt == 3)
}

func TestDoJobName(t *testing.T) {
	t.Parallel()

	j := daemon.NewJob(func(ctx context.Context) error {
		return nil
	})
	requireTrue(t, j.String() == "daemon_test.TestDoJobName.func1")

	jn := daemon.NewJob(func(ctx context.Context) error {
		return nil
	}, jobNameOption)
	requireTrue(t, jn.String() == "job name")
}

func TestDoManagerRetryErr(t *testing.T) {
	t.Parallel()

	var cnt int32

	m := daemon.New(daemon.Retry(3, daemon.StopJob))

	m.Do(ctx, createJob(&cnt, 0, errJob), daemon.WithFreq(time.Second/5))
	m.Wait()
	requireTrue(t, cnt == 3)
}
