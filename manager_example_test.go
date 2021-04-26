package daemon_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"gitoa.ru/go-4devs/daemon"
)

var ErrJob = errors.New("some reason")

func ExampleManager() {
	m := daemon.New()
	j := daemon.NewJob(func(ctx context.Context) error {
		// do some job
		return daemon.StopJob(nil)
	}, daemon.WithName("awesome job"))

	m.Do(context.Background(), j,
		// set frequency run job
		daemon.WithFreq(time.Minute),
		// set delay for first run job
		daemon.WithDelay(time.Second),
		// set handler if run job return err
		daemon.WithHandleErr(func(err error) {
			log.Println(err)
		}),
	)
	m.Wait()
	// Output:
}

func ExampleManager_withClose() {
	m := daemon.New()

	defer func() {
		_ = m.Close()
	}()

	j := daemon.NewJob(func(ctx context.Context) error {
		fmt.Println("do some job;")
		return daemon.StopJob(nil)
	}, daemon.WithName("awesome job"))

	m.Do(context.Background(), j, daemon.WithFreq(time.Microsecond))
	// some blocked process
	time.Sleep(time.Second)
	// Output: do some job;
}

func ExampleManager_withOptions() {
	ctx := context.Background()

	middlewareRun := func(ctx context.Context, next daemon.Run) error {
		fmt.Println("do some before run all job;")

		err := next(ctx)

		fmt.Println("do some after run all job;")

		return err
	}
	middlewareStop := func(ctx context.Context, next daemon.Run) error {
		fmt.Println("do some before close all job;")

		err := next(ctx)

		fmt.Println("do some after close all job;")

		return err
	}
	m := daemon.New(
		daemon.WithRunMiddleware(middlewareRun),
		daemon.WithStopMiddleware(middlewareStop),
		daemon.WithHandleErr(func(err error) {
			// do some if close return err
			log.Println(err)
		}),
	)

	j := daemon.NewJob(func(ctx context.Context) error {
		fmt.Println("do some job;")
		return daemon.StopJob(nil)
	}, daemon.WithName("awesome job"))
	j2 := daemon.NewJob(func(ctx context.Context) error {
		fmt.Println("do some job2;")
		return daemon.StopJob(nil)
	}, daemon.WithName("awesome job2"))

	m.Do(ctx, j, daemon.WithFreq(time.Minute), daemon.WithDelay(time.Second))
	m.Do(ctx, j2, daemon.WithFreq(time.Nanosecond))
	m.Wait()
	// Output:
	// do some before run all job;
	// do some job2;
	// do some after run all job;
	// do some before close all job;
	// do some after close all job;
	// do some before run all job;
	// do some job;
	// do some after run all job;
	// do some before close all job;
	// do some after close all job;
}

func ExampleNewJob() {
	ctx := context.Background()
	m := daemon.New(func(j *daemon.Job) {
		daemon.WithRunMiddleware(func(ctx context.Context, next daemon.Run) error {
			fmt.Printf("running job: %s\n", j)
			return daemon.StopJob(next(ctx))
		})(j)
	})
	j := daemon.NewJob(func(ctx context.Context) error {
		// do some
		return nil
	}, daemon.WithName("my awesome job"))

	m.Do(ctx, j)

	m.Wait()
	// Output: running job: my awesome job
}

func ExampleNewJob_withClose() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m := daemon.New()

	j := daemon.NewJob(func(ctx context.Context) error {
		fmt.Println("do some long job;")
		return daemon.StopJob(nil)
	}, daemon.WithStop(func(ctx context.Context) error {
		fmt.Println("do some close job;")
		return nil
	}))

	m.Do(ctx, j)

	m.Wait()
	// Output:
	// do some long job;
	// do some close job;
}

func ExampleJob_withMiddleware() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m := daemon.New()

	j := daemon.NewJob(
		func(ctx context.Context) error {
			fmt.Println("do some job;")
			return daemon.StopJob(nil)
		},
		daemon.WithStop(func(ctx context.Context) error {
			fmt.Println("do some close job;")
			return nil
		}),
		daemon.WithRunMiddleware(func(ctx context.Context, next daemon.Run) error {
			fmt.Println("do some before run func;")
			err := next(ctx)
			fmt.Println("do some after run func;")
			return err
		}),
		daemon.WithStopMiddleware(func(ctx context.Context, next daemon.Run) error {
			fmt.Println("do some before close func;")
			err := next(ctx)
			fmt.Println("do some after close func;")
			return err
		}),
	)

	m.Do(ctx, j)

	m.Wait()
	// Output:
	// do some before run func;
	// do some job;
	// do some after run func;
	// do some before close func;
	// do some close job;
	// do some after close func;
}

func ExampleJob_option() {
	var cnt int32

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	m := daemon.New()
	j := daemon.NewJob(func(ctx context.Context) error {
		if cnt == 2 {
			return daemon.StopJob(nil)
		}
		atomic.AddInt32(&cnt, 1)
		fmt.Println("do some")
		return nil
	},
		// set freq run job
		daemon.WithFreq(time.Microsecond),
		// set delay to start job
		daemon.WithDelay(time.Nanosecond),
	)

	m.Do(ctx, j)
	m.Wait()
	// Output:
	//do some
	//do some
}

func ExampleNewJob_stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m := daemon.New()

	var i int32

	j := daemon.NewJob(func(ctx context.Context) error {
		atomic.AddInt32(&i, 1)
		fmt.Print("do some:", i, " ")
		return daemon.StopJob(ErrJob)
	})

	m.Do(ctx, j)

	m.Wait()
	// Output: do some:1
}

func ExampleJob_delay() {
	var i int32

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m := daemon.New()
	j := daemon.NewJob(func(ctx context.Context) error {
		if i == 3 {
			return daemon.StopJob(nil)
		}
		atomic.AddInt32(&i, 1)
		fmt.Print("do some:", i, " ")
		return daemon.DelayJob(time.Second/2, ErrJob)
	})

	m.Do(ctx, j)

	m.Wait()
	// Output: do some:1 do some:2 do some:3
}
