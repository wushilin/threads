# Threads

Starting go routine in golang is far too easy, and it is cheap! However, generally it is still a a bad
idea to allow go routine to grow uncontrolled. Hence sometimes you prefer a thread pool concept just like
java. Here you have it

## Install
```
go get github.com/wushilin/threads
```

## Documentation
$ godoc -http=":16666"

Browse http://localhost:16666

## Usage

### Thread Pool API

#### Import

```
import "github.com/wushilin/threads"
```

### Creating a thread pool, with 30 executors and max of 1 million pending jobs
```
thread_pool := threads.NewPool(30, 1000000)

// thread_pool.Start() must be called. Without this, threads won't start processing jobs
thread_pool.Start()
``` 

### Submiting a job and gets a Future
```
fut := thread_pool.Submit(func() interface{} {
  return  1 + 6
})
```

### Wait until the future is ready to retrieve
```
result := fut.GetWait().(int)
fmt.Println("Result of 1 + 6 is", result)

// or if you prefer no blocking
ok, result := fut.GetNoWait()
if ok {
  // result is ready
  fmt.Println("Result of 1 + 6 is", result)
} else [
  fmt.Println("Result is not ready yet")
}

// or if you want to wait for max 3 seconds
ok, result := fut.GetWaitTimeout(3*time.Second)
if ok {
  // result is ready
  fmt.Println("Result of 1 + 6 is", result)
} else [
  fmt.Println("Result is not ready yet")
}
```
### Stop accepting new jobs
```
// once shutdown, you can't re-start it back
thread_pool.Shutdown()
// Wait until all jobs to complete
thread_pool.Wait() 
// after this call, all futures should be able to be retrieved without delay
```

### Getting stats of this pool
```
thread_pool.ActiveCount() // active jobs
thread_pool.PendingCount() // pending count
thread_pool.CompletedCount() //jobs done
```