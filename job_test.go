package daemon

import (
	"context"
	"errors"
	"time"
)

func ExampleNewJob() {
	ctx := context.Background()
	m := NewManager()

	j := NewJob(func(ctx context.Context) error {
		// do some
		return nil
	}, WithJName("my awesome job"))

	m.DoCtx(ctx, j)

	m.Wait()
}

func ExampleNewJob_withClose() {
	ctx := context.Background()
	m := NewManager()

	j := NewJob(func(ctx context.Context) error {
		// do some
		return nil
	}, WithJClose(func(ctx context.Context) error {
		// do some after job stop
		return nil
	}))

	m.DoCtx(ctx, j)

	m.Wait()
}

func ExampleNewJob_withMiddleware() {
	ctx := context.Background()
	m := NewManager()
	mw := NewMiddleware(func(ctx context.Context, next Run) error {
		// do some before run func
		err := next(ctx)
		// do some after run func
		return err
	}, func(ctx context.Context, next Run) error {
		// do some before close func
		err := next(ctx)
		// do some after close func
		return err
	})
	// middleware execute only run func
	mwr := NewMiddleware(func(ctx context.Context, next Run) error {
		return next(ctx)
	}, nil)

	j := NewJob(func(ctx context.Context) error {
		// do some
		return nil
	}, WithJMiddleware(mw, mwr))

	m.DoCtx(ctx, j)

	m.Wait()
}

// all option may replace then DoCtx
func ExampleNewJob_option() {
	ctx := context.Background()
	m := NewManager()

	j := NewJob(func(ctx context.Context) error {
		// do some
		return nil
	},
		// set freq run job
		WithJOption(WithFreq(time.Minute)),
		// set delay to start job
		WithJOption(WithDelay(time.Minute)),
	)

	m.DoCtx(ctx, j)

	m.Wait()
}

func ExampleNewJob_stop() {
	ctx := context.Background()
	m := NewManager()

	j := NewJob(func(ctx context.Context) error {
		// do some
		return StopJob(errors.New("some reason"))
	})

	m.DoCtx(ctx, j)

	m.Wait()
}

// run job after delay once ignore freq
func ExampleNewJob_delay() {
	ctx := context.Background()
	m := NewManager()

	j := NewJob(func(ctx context.Context) error {
		// do some
		return DelayJob(time.Hour, errors.New("some reason"))
	})

	m.DoCtx(ctx, j)

	m.Wait()
}
