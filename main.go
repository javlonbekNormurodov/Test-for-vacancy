package main

import (
	"fmt"
	"sync"
	"time"
)

type Ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT []byte
}

func main() {
	taskCreturer := func(a chan Ttype, wg *sync.WaitGroup) {
		defer wg.Done() // Notify the wait group when this function completes
		for {
			ft := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 {
				ft = "Some error occurred"
			}
			a <- Ttype{cT: ft, id: int(time.Now().Unix())}
		}
	}

	superChan := make(chan Ttype, 10)

	var wg sync.WaitGroup
	wg.Add(1) // Add the task creator to the wait group
	go taskCreturer(superChan, &wg)

	taskWorker := func(a Ttype, wg *sync.WaitGroup, doneTasks chan<- Ttype, undoneTasks chan<- error) {
		defer wg.Done() // Notify the wait group when this function completes

		tt, _ := time.Parse(time.RFC3339, a.cT)
		if tt.After(time.Now().Add(-20 * time.Second)) {
			a.taskRESULT = []byte("task has been successed")
		} else {
			a.taskRESULT = []byte("something went wrong")
		}
		a.fT = time.Now().Format(time.RFC3339Nano)

		time.Sleep(time.Millisecond * 150)

		if string(a.taskRESULT[14:]) == "successed" {
			doneTasks <- a
		} else {
			undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", a.id, a.cT, a.taskRESULT)
		}
	}

	doneTasks := make(chan Ttype)
	undoneTasks := make(chan error)

	go func() {
		// Wait for all task creators to complete before closing superChan
		wg.Wait()
		close(superChan)
	}()

	var taskWorkersWG sync.WaitGroup

	// Start multiple task workers
	for i := 0; i < 5; i++ {
		taskWorkersWG.Add(1)
		go func() {
			defer taskWorkersWG.Done()
			for t := range superChan {
				taskWorker(t, &taskWorkersWG, doneTasks, undoneTasks)
			}
		}()
	}

	go func() {
		// Close the doneTasks and undoneTasks channels once all task workers are done
		taskWorkersWG.Wait()
		close(doneTasks)
		close(undoneTasks)
	}()

	// Collect the results
	result := make(map[int]Ttype)
	errs := []error{}

	for r := range doneTasks {
		result[r.id] = r
	}

	for r := range undoneTasks {
		errs = append(errs, r)
	}

	// Print errors
	fmt.Println("Errors:")
	for _, r := range errs {
		fmt.Println(r)
	}

	// Print done tasks
	fmt.Println("Done tasks:")
	for _, r := range result {
		fmt.Println(r.id)
	}
}
