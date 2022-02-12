package threads

import (
	"sync"
	"sync/atomic"
	"time"
  "reflect"
  future "github.com/wushilin/future"
)

var globalMutex sync.Mutex

// Represent a Thread Pool Object
type ThreadPool struct {
	limit            int
	jobs             chan any
	active_count     int32
	completion_count int64
	wg               *sync.WaitGroup
	startedAt        time.Time
  shutdownAt       time.Time
}

// Internal concept of a Job and its produced result
type Job[T any] struct {
	Jobf   func() T
	Result future.Future[T]
}

// Create a new ThreadPool. threads is the Max concurrent thread (go routines)
// max_pending_jobs is max number of pending jobs. If the pending jobs is full
// Calling submit will be blocked until a vacancy is available
func NewPool(threads int, max_pending_jobs int) *ThreadPool {
	return &ThreadPool{limit: threads, jobs: make(chan any, max_pending_jobs), active_count: 0, completion_count: 0, wg: new(sync.WaitGroup)}
}

func launch(f func()) {
	go f()
}
// Starts the worker threads. Number of threads is represented by pool's threads configuration
func (v *ThreadPool) Start() {
	globalMutex.Lock()
		defer globalMutex.Unlock()
		if !v.IsStarted() {
			for i := 0; i < v.limit; i++ {
				v.wg.Add(1)
					launch(func() {
							defer v.wg.Done()
							for nextr := range v.jobs {
							atomic.AddInt32(&v.active_count, 1)
							rv := reflect.Indirect(reflect.ValueOf(nextr))
							jobResult := rv.FieldByName("Jobf").Call([]reflect.Value{})
							if len(jobResult) != 1 {
							panic("Task did not have a single value result")
							}
							rv.FieldByName("Result").MethodByName("Set").Call(jobResult)
							atomic.AddInt32(&v.active_count, -1)
							atomic.AddInt64(&v.completion_count, 1)
							}
							})
			}
			v.startedAt = time.Now()
		} else {
			panic("Why you want to start your ThreadPool twice?")
		}
}

// Returns how many threads are working on job currently
func (v *ThreadPool) ActiveCount() int {
	return int(v.active_count)
}

func (v *ThreadPool) IsStarted() bool {
	return !v.startedAt.IsZero()
}

// Returns when the thread pool gets started
func (v *ThreadPool) StartedTime() time.Time {
	return v.startedAt
}

func (v *ThreadPool) ShutdownTime() time.Time {
	return v.shutdownAt
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
// You can't shutdown more than once, sorry
func (v *ThreadPool) Shutdown() {
	close(v.jobs) //now submission will panic
  v.shutdownAt = time.Now()
}

func (v *ThreadPool) IsShutdown() bool {
	return !v.shutdownAt.IsZero()
}

// Wait until all jobs are processed. after this, All previously returned future should be ready for retrieval
// Must call Shutdown() first or Wait() will block forever
func (v *ThreadPool) Wait() {
  if ! v.IsShutdown() {
		panic("Possible deadlock: The ThreadPool has not been shutdown yet!")
  }
	v.wg.Wait()
}

func SubmitTasks[T any](pool *ThreadPool, jobs []func() T) *FutureGroup[T] {
	result := make([]future.Future[T], len(jobs))
	for idx, nj := range jobs {
		result[idx] = SubmitTask[T](pool, nj)
	}
	return NewFutureGroup[T](result, pool)
}

func SubmitTask[T any](pool *ThreadPool, task func() T) future.Future[T] {
  if ! pool.IsStarted() {
    panic("Possible deadlock: The pool is not started yet")
  }
  var zv T
	result := future.NewPendingFuture[T](zv)
  nj := &Job[T]{task, result}
  pool.jobs <- nj
  return result
}

type FutureGroup[T any] struct {
	futures []future.Future[T]
	pool    *ThreadPool
	flags   []bool
}

func NewFutureGroup[T any](futures []future.Future[T], pool *ThreadPool) *FutureGroup[T] {
	flags := make([]bool, len(futures))

	return &FutureGroup[T]{futures, pool, flags}
}
func (v *FutureGroup[T]) Count() int64 {
	return int64(len(v.futures))
}

func (v *FutureGroup[T]) Futures() []future.Future[T] {
	result := make([]future.Future[T], len(v.futures))

	for i := 0; i < len(v.futures); i++ {
		result[i] = v.futures[i]
	}

	return result
}

func (v *FutureGroup[T]) ThreadPool() *ThreadPool {
	return v.pool
}

func (v *FutureGroup[T]) ReadyCount() int64 {
	var sum int64 = 0
	for idx, nf := range v.futures {
		if v.flags[idx] {
			sum++
			continue
		}
		ready, _ := nf.GetNow()
		if ready {
			sum++
			v.flags[idx] = true
		}
	}
	return sum
}

func (v *FutureGroup[T]) IsAllReady() bool {
	return v.ReadyCount() == v.Count()
}

func (v *FutureGroup[T]) WaitTimeOut(timeout time.Duration) (bool, []any) {
	resultchan := make(chan []any, 1)
	launch(func() {
		resultchan <- v.WaitAll()
		close(resultchan)
	})
	select {
	case result := <-resultchan:
		return true, result
	case <-time.After(timeout):
		return false, nil
	}
}

func (v *FutureGroup[T]) WaitAll() []any {
	result := make([]any, len(v.futures))
	for idx, nf := range v.futures {
		result[idx] = nf.GetWait()
	}
	return result
}

// Do a list of jobs in parallel, and return the List of futures immediately
// This will create as many threads as possible
func ParallelDo[T any](jobs []func() T) *FutureGroup[T] {
	return ParallelDoWithLimit[T](jobs, len(jobs))
}

// Do a list of jobs in parallel, and return the List of futures immediately
func ParallelDoWithLimit[T any](jobs []func() T, nThreads int) *FutureGroup[T] {
	tp := NewPool(nThreads, len(jobs))
	tp.Start()
	defer func() {
		tp.Shutdown()
		//tp.Wait()
	}()
	result := make([]future.Future[T], len(jobs))
	for idx, nj := range jobs {
		result[idx] = SubmitTask[T](tp, nj)
	}
	return NewFutureGroup[T](result, tp)
}
