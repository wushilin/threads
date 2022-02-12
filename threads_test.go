package threads

import (
	"fmt"
	"testing"
	"time"
  future "github.com/wushilin/future"
)

func TestFuture(t *testing.T) {
  tp := NewPool(10, 100000)
	tasks := make([]func() int, 10)

	for i := 0; i < len(tasks); i++ {
		tasks[i] = sleeper
	}
	fmt.Printf("About to start\n")
  tp.Start()
	fg := SubmitTasks(tp, tasks)
	fmt.Printf("FutureGroup is %v\n", fg)
	for {
		fmt.Printf("Count: %d\n", fg.Count())
		fmt.Printf("ReadyCount: %d\n", fg.ReadyCount())
		ready, result := fg.WaitTimeOut(100 * time.Millisecond)
		fmt.Println(ready, result)
		fmt.Printf("Active count: %d\n", fg.ThreadPool().ActiveCount())
		fmt.Printf("Pending jobs: %d\n", fg.ThreadPool().PendingCount())
		fmt.Printf("Completed jobs: %d\n", fg.ThreadPool().CompletedCount())
		if ready {
			break
		}
	}

	fmt.Println(fg.IsAllReady())
	fmt.Println("About to wait")
  tp.Shutdown()
	tp.Wait()
	fmt.Println("Wait done")

	futInstant := future.InstantFutureOf(5)
	future.Chain(futInstant, mapadd).Then(printer[int])
}

func mapadd(i int) int {
	return i + 1
}

func sleeper() int {
	time.Sleep(1 * time.Second)
	return 5
}

func printer[T any](i T) {
	fmt.Printf("You got %v\n", i)
}
