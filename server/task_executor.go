package main

import (
	"log"
)

const TASK_EXECUTOR_QUEUE_SIZE = 100

func NewTaskExecutor(server *Server) *TaskExecutor {
	ex := &TaskExecutor{}
	ex.queue = make(chan Task, TASK_EXECUTOR_QUEUE_SIZE)
	ex.server = server
	return ex
}

type Task interface {
	Run(ex *TaskExecutor)
}

type TaskExecutor struct {
	queue  chan Task
	server *Server
}

func (ex *TaskExecutor) Submit(task Task) {
	ex.queue <- task
}

func (ex *TaskExecutor) Start() {
	go func() {
		exit := false
		for !exit {
			exit = ex.run()
		}
	}()
}

func (ex *TaskExecutor) run() (exit bool) {

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Println("TaskExecutor Run Error: ", err)
			} else {
				log.Println("TaskExecutor Run Panic: ", r)
			}
		}
		exit = false
	}()

	for {
		// Get one task from queu
		task := <-ex.queue
		// Execute
		task.Run(ex)
	}
}
