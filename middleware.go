package daemon

import "context"

//Middleware midleware to set manager or job
type Middleware struct {
	close Handle
	run   Handle
}

//NewMiddleware create new Middleware
func NewMiddleware(run Handle, close Handle) Middleware {
	return Middleware{run: run, close: close}
}

//RetryMiddleware set retry job and change return after max retry
func RetryMiddleware(max uint8, handleRetry func(err error) error) Middleware {
	var retry uint8
	return NewMiddleware(func(ctx context.Context, next Run) error {
		err := next(ctx)
		if err != nil {
			retry++
		}
		if retry >= max {
			return handleRetry(err)
		}
		return err
	}, nil)
}
