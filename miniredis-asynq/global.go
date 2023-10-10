package main

import "github.com/hibiken/asynq"

type GlobalApplicationState struct {
	srv         *asynq.Server
	inspector   *asynq.Inspector
	scheduler   *asynq.Scheduler
	queueclient *asynq.Client
	mux         *asynq.ServeMux
}

type RedisReady struct{}

const redisAddr = "localhost:6379"
