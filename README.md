# I have re-written this with Generics, Yay!
Thanks to go generics, the code does not use interface{} now!

However, due to Golang generics limitation (methods can't introduce new type parameter), so instead of 
```go
pool.Submit(func() interface{})
```
We now have to use 
```go
threads.SubmitTask(pool, func() T)
```

I personally don't think this is a big deal


# Threads

Starting go routine in golang is far too easy, and it is cheap! However, generally it is still a a bad
idea to allow go routine to grow uncontrolled. Hence sometimes you prefer a thread pool concept just like
java. Here you have it

## Install
```go
go get github.com/wushilin/threads
```

## Documentation
$ godoc -http=":16666"

Browse http://localhost:16666

## Usage

### Thread Pool API

#### Import

```go
import "github.com/wushilin/threads"
```

### Creating a thread pool, with 30 executors and max of 1 million pending jobs
```go
var thread_pool *threads.ThreadPool = threads.NewPool(30, 1000000)
// Note that if pending job is more than 1000000, the new submission (call to Submit) will be blocked
// until the job queue has some space.

// thread_pool.Start() must be called. Without this, threads won't start processing jobs
// You now can't submit jobs before pool is started because it may cause dead lock if the buffer is not enough.
thread_pool.Start()

// After thread_pool is started, there is 30 go routines in background, processing jobs


``` 

### Submiting a job and gets a Future
```go
var fut future.Future = thread_pool.SubmitTask(thread_poo, func() int {
  return  1 + 6
})

// Here, submited func that returns a value. The func will be executed by a backend processor
// where there is free go routine. The submission returns a *threads.Future, which can be used
// to retrieve the returned value from the func. 
// e.g. 
// resultInt := fut.GetWait() // <= resultInt will be 7, and type is int. Thanks to genercs in go
```

### Wait until the future is ready to be retrieve
```go
result := fut.GetWait() // <= result will be 7
fmt.Printf("Result of 1 + 6 is %d", result)
// Wait until it is run and result is ready

// or if you prefer no blocking, call returns immediately, but may contain no result
ok, result := fut.GetNow()
if ok {
  // result is ready
  fmt.Println("Result of 1 + 6 is", result) // <= result will be 7
} else {
  fmt.Println("Result is not ready yet")
}

// or if you want to wait for max 3 seconds
ok, result := fut.GetTimeout(3*time.Second)
if ok {
  // result is ready
  fmt.Println("Result of 1 + 6 is", result) // <= result will be 7
} else {
  fmt.Println("Result is not ready yet") // <= timed out after 3 seconds
}
```
### Stop accepting new jobs
```go
// once shutdown, you can't re-start it back
thread_pool.Shutdown()
// Now thread_pool can't submit new jobs. All existing submited jobs will be still processed
// The future previous returned will still materialize

// Wait until all jobs to complete. Calling Wait() on non-shutdown thread pool will be blocked forever
thread_pool.Wait() 
// You can't call Wait() before you call Shutdown because it may cause dead lock
// after this call, all futures should be able to be retrieved without delay
// You can safely disregard this thread_pool after this call. It is useless anyway
```

### Getting stats of this pool
```go
thread_pool.ActiveCount() // active jobs - being executed right now
thread_pool.PendingCount() // pending count - not started yet
thread_pool.CompletedCount() //jobs done - result populated already
```

### Convenient wrapper to do multiple tasks in parallel
```go
jobs := make([]func() int, 60)
//... populate the jobs with actual jobs
// This will start as many threads as possible to run things in parallel
var fg *threads.FutureGroup = threads.ParallelDo(jobs)

// This will start at most 10 threads for parallel processing
var fg *threads.FutureGroup = threads.ParallelDoWithLimit(jobs, 10)

// retrieve futures, wait for all and get result!
var results []int = fg.WaitAll()

// If you prefer more flexible handling... - you get a copy of the array
var []future.Future futures = fg.Futures()

```

# Interesting future concepts
see github.com/wushilin/future
```go
fut.Then(print) => print function is called with argument of Future's value, when value become available

fut.Then(print).Then(save) => multiple then functions can be called

fut := SubmitTask(thread_pool, func() int {
	return 5
})
fut2 := Chain(fut, func(i int) string {
	return fmt.Sprintf("Student #%d", i)
}

//fut2 is a future of "string" instead of "int" now. 

fut2.Then(print)
// print fut2 when it is available

fut3 := future.DelayedFutureOf("hello how are you", 3 * time.Second) => fut3 is available after 3 seconds
```

# Future Group now supports
```
FutureGroup.Count() // Count the number of futures in the group
FutureGroup.Futures() // Get a copy of futures (not the underlying future directly)
FutureGroup.ReadyCount() // Check how many of futures are ready
FutureGroup.IsAllReady() // Test if all results are present (non-blocking)
FutureGroup.ThreadPool() // returns original thread pool that produced the future group. You may want to call its Wait() methods (but usually not necessary)
```
