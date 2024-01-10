package main

import (
	"fmt"
	"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// Task type represents a task struct
type Task struct {
	ID           int
	CreatedTime  time.Time
	FinishedTime time.Time
	Result       string
	Error        error
}

// TaskResult type represents a task result struct
type TaskResult struct {
	Task  Task
	Error error
}

func main() {
	createTasks := func(taskChannel chan<- Task, count int) {
		for i := 0; i < count; i++ {
			currentTime := time.Now()
			var taskError error
			if currentTime.Nanosecond()%2 > 0 {
				taskError = fmt.Errorf("some error occurred")
			}
			taskChannel <- Task{ID: int(currentTime.Unix()), CreatedTime: currentTime, Error: taskError}
			time.Sleep(100 * time.Millisecond)
		}
		close(taskChannel)
	}

	const totalTasks = 10
	taskChannel := make(chan Task, totalTasks)
	resultChannel := make(chan TaskResult, totalTasks)
	var workers sync.WaitGroup

	go createTasks(taskChannel, totalTasks)

	processTask := func(task Task) {
		defer workers.Done()
		var result TaskResult
		result.Task = task
		if task.Error == nil && time.Since(task.CreatedTime) < 20*time.Second {
			result.Task.Result = "task has been succeeded"
		} else {
			result.Error = fmt.Errorf("task id %d, time %s, error %v", task.ID, task.CreatedTime, task.Error)
		}
		result.Task.FinishedTime = time.Now()
		resultChannel <- result
	}

	go func() {
		for task := range taskChannel {
			workers.Add(1)
			go processTask(task)
		}
		workers.Wait()
		close(resultChannel)
	}()

	results := map[int]Task{}
	errors := []error{}

	for result := range resultChannel {
		if result.Error != nil {
			errors = append(errors, result.Error)
		} else {
			results[result.Task.ID] = result.Task
		}
	}

	fmt.Println("Errors:")
	for _, e := range errors {
		fmt.Println(e)
	}

	fmt.Println("Done tasks:")
	for id, task := range results {
		fmt.Printf("Task ID: %d, Result: %s\n", id, task.Result)
	}
}
