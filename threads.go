package threads

import (
	"sync"
	"sync/atomic"
	"time"
)

// Represent a Thread Pool Object
type ThreadPool struct {
	limit            int
	jobs             chan *Job
	active_count     int32
	completion_count int64
	wg               *sync.WaitGroup
	started          time.Time
}

// Represents a Future result object
type Future struct {
	result interface{}
	signal chan bool // if read channel returned, result is ready
}

// Represent a job that returns something for future retrieval. Could be nil
type JobFunc func() interface{}

// Internal concept of a Job and its produced result
type Job struct {
	jobf   JobFunc
	result *Future
}

// Create a new ThreadPool. threads is the Max concurrent thread (go routines)
// max_pending_jobs is max number of pending jobs. If the pending jobs is full
// Calling submit will be blocked until a vacancy is available
func NewPool(threads int, max_pending_jobs int) *ThreadPool {
	return &ThreadPool{limit: threads, jobs: make(chan *Job, max_pending_jobs), active_count: 0, completion_count: 0, wg: new(sync.WaitGroup)}
}

// Starts the worker threads. Number of threads is represented by pool's threads configuration
func (v *ThreadPool) Start() {
	for i := 0; i < v.limit; i++ {
		v.wg.Add(1)
		go func() {
			defer v.wg.Done()
			for next := range v.jobs {
				atomic.AddInt32(&v.active_count, 1)
				result := next.jobf()
				next.result.updateResult(result)
				atomic.AddInt32(&v.active_count, -1)
				atomic.AddInt64(&v.completion_count, 1)
			}
		}()
	}
	v.started = time.Now()
}

// Returns how many threads are working on job currently
func (v *ThreadPool) ActiveCount() int {
	return int(v.active_count)
}

// Returns when the thread pool gets started
func (v *ThreadPool) StartedTime() time.Time {
	return v.started
}

// How many jobs are still in the queue (not started)
func (v *ThreadPool) PendingCount() int {
	return len(v.jobs)
}

// How many jobs has been completed
func (v *ThreadPool) CompletedCount() int64 {
	return v.completion_count
}

// Stop accepting new jobs. After this call is called, future calls to Submit will panic
func (v *ThreadPool) Shutdown() {
	close(v.jobs) //now submission will panic
}

// Wait until all jobs are processed (after this, All previously returned future should be ready for retrieval
func (v *ThreadPool) Wait() {
	v.wg.Wait()
}

// Submit a job and return a Future value can be retrieved later sync or async
func (v *ThreadPool) Submit(j JobFunc) *Future {
	if j == nil {
		panic("Can't submit nill function")
	}
	result := &Future{nil, make(chan bool, 1)}
	nj := &Job{j, result}
	v.jobs <- nj
	return result
}

func (v *Future) updateResult(result interface{}) {
	v.result = result
	v.signal <- true
	close(v.signal)
}

// Get the future value without wait. bool value is whether this retrieve did retrieve something, the interface{} value
// is the actual future result
func (v *Future) GetNoWait() (bool, interface{}) {
	return v.GetWaitTimeout(0 * time.Second)
}

// Synchronously retrieve the future's value. It will block until the value is available
func (v *Future) GetWait() interface{} {
	<-v.signal
	return v.result
}

// Retrieve the futures value, with a timeout. The bool value represent whether this retrieval did succeed
func (v *Future) GetWaitTimeout(t time.Duration) (bool, interface{}) {
	select {
	case <-v.signal:
		return true, v.result
	case <-time.After(t):
		return false, nil
	}
}