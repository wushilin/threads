package threads

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	future "github.com/wushilin/future"
)

func TestFuture(t *testing.T) {
	tp := NewPool(10, 1000000)
	tasks := make([]func() int, 1000)

	for i := 0; i < len(tasks); i++ {
		tasks[i] = sleeper
	}
	fmt.Printf("About to start\n")
	tp.Start()
	fg := SubmitTasks(tp, tasks)
	go func() {
		for {
			ok, result := fg.WaitTimeOut(time.Second)
			if !ok {
				fmt.Println("fg.WaitTimeout didn't succeed")
			} else {
				fmt.Printf("fg.WaitTimeout did work! %v\n", result)
				return
			}
		}
	}()
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

var r = rand.New(rand.NewSource(99))

func sleeper() int {
	time.Sleep(time.Duration((r.Int31() % 100)) * time.Millisecond)
	return 1
}

func printer[T any](i T) {
	fmt.Printf("You got %v\n", i)
}
