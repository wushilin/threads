package threads

import (
	"fmt"
	"testing"
	"time"
)

func TestFuture(t *testing.T) {
	/*
		future1 := FutureOf(sleeper)
		fmt.Println(future1.GetNoWait())
		fmt.Println(future1.GetNoWait())

		future2 := future1.ThenMap(mapadd)
		future2.Then(printer)
		future3 := future2.ThenMap(mapadd)
		future3.Then(printer)
		fmt.Println(future2.GetNoWait())

	*/
	tasks := make([]JobFunc, 10000)

	for i := 0; i < len(tasks); i++ {
		tasks[i] = sleeper
	}
	fmt.Printf("About to start\n")
	fg := ParallelDoWithLimit(tasks, 500)
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
	fg.ThreadPool().Wait()
	fmt.Println("Wait done")

	futInstant := Instantly(5)
	futInstant.ThenMap(mapadd).Then(printer)
}

func mapadd(i interface{}) interface{} {
	return i.(int) + 1
}

func sleeper() interface{} {
	time.Sleep(1 * time.Second)
	return 5
}

func printer(i interface{}) {
	fmt.Printf("You got %v\n", i)
}
