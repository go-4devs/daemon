# daemon
[![Build Status](https://travis-ci.org/go-4devs/daemon.svg?branch=master)](https://travis-ci.org/go-4devs/daemon)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-4devs/daemon)](https://goreportcard.com/report/github.com/go-4devs/daemon)

the logic to manage the background processes [GoDoc](https://godoc.org/github.com/go-4devs/daemon)
## install

```bash
$ go get github.com/go-4devs/daemon@latest
```

## basic usage

```go
package main

import (
	"context"
	"log"
	"time"
	
	"github.com/go-4devs/daemon"
)


func main() {
    ctx := context.Background()
	logErr := daemon.NewMiddleware(func(ctx context.Context, next daemon.Run) error {
	    err := next(ctx)
	    if err != nil{	    	
	       log.Println(err)
	    }
	    return err
	},nil)
    manager := daemon.NewManager(daemon.WithMMiddleware(logErr))
    
    job := daemon.NewJob(func(ctx context.Context) error {
        // my some job
        return nil
    })
    //do job immediately and run once a minute
    manager.DoCtx(ctx, job, daemon.WithFreq(time.Minute))
    
    //do another job after 1 hour and run once a second
    manager.DoCtx(ctx, job, daemon.WithDelay(time.Hour), daemon.WithFreq(time.Second))
    
    //single run job
    jobWithClose := daemon.NewJob(func(ctx context.Context) error {
        // my some logic
        return daemon.StopJob(nil)
    }, daemon.WithJClose(func(ctx context.Context) error {
    	// do some logic after job stop
        return nil
    }))
   
    manager.DoCtx(ctx, jobWithClose)
    
    // wait all jobs
    manager.Wait()
}
```
