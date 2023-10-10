package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/davidroman0O/seigyo"
)

func main() {

	// Create a channel to receive OS signals.
	sigCh := make(chan os.Signal, 1)

	// Notify sigCh when receiving SIGINT or SIGTERM signals.
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	controller := seigyo.New[GlobalApplicationState](GlobalApplicationState{})
	controller.RegisterProcess("redis", seigyo.ProcessConfig[GlobalApplicationState]{
		Process: &RedisProcess{},
	})
	controller.RegisterProcess("worker", seigyo.ProcessConfig[GlobalApplicationState]{
		Process: &WorkerProcess{},
	})
	controller.RegisterProcess("api", seigyo.ProcessConfig[GlobalApplicationState]{
		Process: &ApiProcess{},
	})

	// Start all processes.
	errCh := controller.Start()

	go func() {
		// Wait for an OS signal.
		<-sigCh
		controller.Stop()
	}()

	for err := range errCh {
		fmt.Println(err)
		controller.Stop()
	}

	// Stop all processes.
	controller.Stop()
}
