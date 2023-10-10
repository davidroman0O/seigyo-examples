package main

import (
	"context"
	"time"

	"github.com/alicebob/miniredis/v2"
)

type RedisProcess struct {
	mini *miniredis.Miniredis
}

func (r *RedisProcess) Init(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{})) error {

	r.mini = miniredis.NewMiniRedis()

	return nil
}

func (r *RedisProcess) Run(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{}),
	shutdownCh chan struct{},
	errCh chan<- error) error {

	go func() {
		if err := r.mini.StartAddr(":6379"); err != nil {
			errCh <- err
		}
	}()

	if err := r.mini.Set("foo", "bar"); err != nil {
		errCh <- err
	}

	time.Sleep(time.Second * 5)
	sender("worker", RedisReady{})
	sender("api", RedisReady{})

	<-shutdownCh

	r.mini.Close()

	return nil
}

func (r *RedisProcess) Deinit(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{})) error {

	return nil
}

func (r *RedisProcess) Received(
	pid string,
	data interface{}) error {
	return nil
}
