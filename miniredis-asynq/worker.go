package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

type WorkerProcess struct {
	readyRedis bool
}

func (r *WorkerProcess) Init(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{})) error {

	for !r.readyRedis {
		time.Sleep(time.Millisecond * 150)
	}

	stateMutator(func(gas GlobalApplicationState) GlobalApplicationState {

		gas.srv = asynq.NewServer(
			asynq.RedisClientOpt{
				Addr: redisAddr,
			},
			asynq.Config{
				// Specify how many concurrent workers to use
				Concurrency: 10,
				Queues: map[string]int{
					"queue": 10,
				},
				ShutdownTimeout: time.Second * 1,
			},
		)

		gas.inspector = asynq.NewInspector(
			asynq.RedisClientOpt{
				Addr: redisAddr,
			},
		)

		gas.scheduler = asynq.NewScheduler(
			asynq.RedisClientOpt{
				Addr: redisAddr,
			},
			&asynq.SchedulerOpts{},
		)

		gas.queueclient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

		gas.scheduler.Register("@every 1s", asynq.NewTask("cron", nil, asynq.Queue("queue")))

		// mux maps a type to a handler
		gas.mux = asynq.NewServeMux()
		gas.mux.HandleFunc("cron", func(ctx context.Context, t *asynq.Task) error {
			fmt.Println("cron")
			return nil
		})

		return gas
	})

	return nil
}

func (r *WorkerProcess) Run(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{}),
	shutdownCh chan struct{},
	errCh chan<- error) error {

	go stateGetter().scheduler.Run()

	go func() {
		if err := stateGetter().srv.Run(stateGetter().mux); err != nil {
			errCh <- err
		}
	}()

	<-shutdownCh

	stateGetter().scheduler.Shutdown()
	stateGetter().srv.Stop()
	stateGetter().srv.Shutdown()

	return nil
}

func (r *WorkerProcess) Deinit(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{})) error {

	return nil
}

func (r *WorkerProcess) Received(
	pid string,
	data interface{}) error {

	switch data.(type) {
	case RedisReady:
		r.readyRedis = true
	}

	return nil
}
