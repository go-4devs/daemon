package daemon

import "context"

// Retry set retry job and change return after max retry.
func Retry(max uint8, handleRetry func(err error) error) Option {
	var retry uint8

	return func(j *Job) {
		WithRunMiddleware(func(ctx context.Context, next Run) error {
			if err := next(ctx); err != nil {
				retry++
				if retry >= max {
					return handleRetry(err)
				}
				return err
			}
			retry = 0
			return nil
		})(j)
	}
}

// RunOnce run once and stopped job.
func RunOnce() Option {
	return func(j *Job) {
		WithRunMiddleware(func(ctx context.Context, next Run) error {
			return StopJob(next(ctx))
		})(j)
	}
}
