package daemon

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	ctx           = context.Background()
	errJob        = errors.New("error job")
	jobName       = "job name"
	jobNameOption = WithJName(jobName)
)

func ExampleNewManager() {
	m := NewManager()
	j := NewJob(func(ctx context.Context) error {
		// do some job
		return nil
	}, WithJName("awesome job"))

	m.DoCtx(context.Background(), j,
		// set frequency run job
		WithFreq(time.Minute),
		// set delay for first run job
		WithDelay(time.Second),
		// set handler if run job return err
		WithHandleErr(func(err error) {
			log.Println(err)
		}),
	)
	m.Wait()
}

func ExampleNewManager_withClose() {
	m := NewManager()
	defer func() {
		_ = m.Close()
	}()

	j := NewJob(func(ctx context.Context) error {
		// do some job
		return nil
	}, WithJName("awesome job"))

	m.DoCtx(context.Background(), j, WithFreq(time.Minute))
	exDone := make(chan struct{})
	// some blocked process
	<-exDone
}

func ExampleNewManager_withOptions() {

	middleware := NewMiddleware(func(ctx context.Context, next Run) error {
		// do some before run all job
		err := next(ctx)
		// do some after run all job
		return err
	}, func(ctx context.Context, next Run) error {
		// do some before close job
		err := next(ctx)
		// do some after close job
		return err
	})
	m := NewManager(
		WithMMiddleware(middleware),
		WithHandleCloseErr(func(err error) {
			// do some if close return err
			log.Println(err)
		}),
	)

	j := NewJob(func(ctx context.Context) error {
		// do some job
		return nil
	}, WithJName("awesome job"))

	m.DoCtx(context.Background(), j, WithFreq(time.Minute))
}

func createJob(count *int, d time.Duration, err error, opts ...JOption) Job {
	opts = append(opts, jobNameOption, WithJOption())
	return NewJob(func(ctx context.Context) error {
		if d > 0 {
			<-time.After(d)
		}
		*count++
		return err
	}, opts...)
}

func TestDoJobSuccess(t *testing.T) {
	t.Parallel()

	m := NewManager()
	cnt := 0
	m.DoCtx(ctx, createJob(&cnt, 0, nil), WithFreq(time.Second/3))
	cnt2 := 0
	runm := 0
	clm := 0
	mw := NewMiddleware(func(ctx context.Context, next Run) error {
		runm++
		return next(ctx)
	}, func(ctx context.Context, next Run) error {
		clm++
		return next(ctx)
	})
	m.DoCtx(ctx, createJob(&cnt2, 0, nil, WithJMiddleware(mw, mw, mw)), WithFreq(time.Second/10))
	time.AfterFunc(time.Second, func() {
		require.Nil(t, m.Close())
	})
	m.Wait()

	require.True(t, cnt > 2)
	require.True(t, cnt2 >= 10)
	require.True(t, cnt2*3 == runm)
	require.True(t, clm == 3)
}

func TestDoJobStop(t *testing.T) {
	t.Parallel()

	m := NewManager()
	cnt := 0
	m.DoCtx(ctx, createJob(&cnt, 0, StopJob(nil)), WithFreq(time.Nanosecond))
	m.Wait()
	require.Equal(t, int(1), cnt)
}

func TestDoJobDelay(t *testing.T) {
	t.Parallel()

	m := NewManager()
	cnt := 0
	m.DoCtx(ctx, createJob(&cnt, 0, DelayJob(time.Second/3, nil)), WithFreq(time.Second))
	time.AfterFunc(time.Second, func() {
		require.Nil(t, m.Close())
	})
	m.Wait()
	require.True(t, cnt > 2)
}

func TestDoJobSkipErr(t *testing.T) {
	t.Parallel()

	m := NewManager()
	cnt := 0
	m.DoCtx(ctx, createJob(&cnt, 0, errJob), WithFreq(time.Second/3))
	time.AfterFunc(time.Second, func() {
		require.Nil(t, m.Close())
	})
	m.Wait()
	require.True(t, cnt > 2)
}

func TestDoJobRetryErr(t *testing.T) {
	t.Parallel()

	m := NewManager()
	cnt := 0
	m.DoCtx(ctx, createJob(&cnt, time.Millisecond, errJob, WithJMiddleware(RetryMiddleware(3, StopJob))), WithFreq(time.Nanosecond))
	m.Wait()
	require.True(t, cnt == 3)
}

func TestDoJobName(t *testing.T) {
	t.Parallel()
	j := NewJob(func(ctx context.Context) error {
		return nil
	})
	require.Equal(t, "daemon.TestDoJobName.func1", j.String())

	jn := NewJob(func(ctx context.Context) error {
		return nil
	}, jobNameOption)
	require.Equal(t, "job name", jn.String())
}

func TestDoManagerRetryErr(t *testing.T) {
	t.Parallel()

	m := NewManager(WithMMiddleware(RetryMiddleware(3, StopJob)))
	cnt := 0
	m.DoCtx(ctx, createJob(&cnt, 0, errJob), WithFreq(time.Second/5))
	m.Wait()
	require.True(t, cnt == 3)
}
