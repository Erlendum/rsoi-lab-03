package main

import (
	"context"
	"github.com/Erlendum/rsoi-lab-03/internal/gateway/manager"
	"os"
	"os/signal"
	"syscall"
)

type root interface {
	Register(ctx context.Context) error
	Resolve(ctx context.Context, shutdown chan os.Signal) os.Signal
	Release(ctx context.Context, signal os.Signal)
}

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	var r root
	r = manager.NewRoot()

	err := r.Register(context.Background())
	if err != nil {
		os.Exit(1)
	}

	s := r.Resolve(context.Background(), shutdown)

	r.Release(context.Background(), s)
}
