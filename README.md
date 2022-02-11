# I will rewrite this with Generics when Go1.18 is realeased. Stay tuned!


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
thread_pool.Start()
// After thread_pool is started, there is 30 go routines in background, processing jobs


``` 

### Submiting a job and gets a Future
```go
var fut *threads.Future = thread_pool.Submit(func() interface{} {
  return  1 + 6
})

// Here, submited func that returns a value. The func will be executed by a backend processor
// where there is free go routine. The submission returns a *threads.Future, which can be used
// to retrieve the returned value from the func. 
// e.g. 
// resultInt := fut.GetWait().(int) // <= resultInt will be 7
```

### Wait until the future is ready to be retrieve
```go
result := fut.GetWait().(int) // <= result will be 7
fmt.Println("Result of 1 + 6 is", result)
// Wait until it is run and result is ready

// or if you prefer no blocking, call returns immediately, but may contain no result
ok, result := fut.GetNoWait()
if ok {
  // result is ready
  fmt.Println("Result of 1 + 6 is", result) // <= result will be 7
} else [
  fmt.Println("Result is not ready yet")
}

// or if you want to wait for max 3 seconds
ok, result := fut.GetWaitTimeout(3*time.Second)
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
jobs := make([]func() interface{}, 60)
//... populate the jobs with actual jobs
// This will start as many threads as possible to run things in parallel
var fg *threads.FutureGroup = threads.ParallelDo(jobs)

// This will start at most 10 threads for parallel processing
var fg *threads.FutureGroup = threads.ParallelDoWithLimit(jobs, 10)

// retrieve futures, wait for all and get result!
var results[]interface{} = fg.WaitAll()

// If you prefer more flexible handling... - you get a copy of the array
var []*threads.Future futures = fg.Futures()

```

# Chain futures, to get futures from futures

```
func adder(i interface{}) interface{} {
  return i.(int) + 1
}

func sleeper() interface{} {
  time.Sleep(1 * time.Second)
  return 5
}

func print_value(i interface{}) {
  fmt.Println(i)
}

future1 := FutureOf(sleeper).ThenMap(adder) // Non blocking, gets the future immediately: future1 materialize when sleeper return, and final value will be 6 
//(sleeper returned value will be then added 1)
future1.Then(print_value) // Non blocking, print_value is immediately run after future1 is materialized
```

# Change FutureGroup from struct to pointer, to save possible large array copy (go default uses pass by value)
All code that was using FutureGroup, now should use *FutureGroup

# Future Group now supports
```
FutureGroup.Count() // Count the number of futures in the group
FutureGroup.Futures() // Get a copy of futures (not the underlying future directly)
FutureGroup.ReadyCount() // Check how many of futures are ready
FutureGroup.IsAllReady() // Test if all results are present (non-blocking)
FutureGroup.ThreadPool() // returns original thread pool that produced the future group. You may want to call its Wait() methods (but usually not necessary)
```
# Added Instantly method

```
threads.InstantFuture(5) => Get a Future that materializes instantly with value 5
```
