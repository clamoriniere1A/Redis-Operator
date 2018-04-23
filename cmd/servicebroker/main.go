package main

import (
	"context"
	goflag "flag"
	"os"
	"runtime"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"github.com/amadeusitgroup/redis-operator/pkg/broker/config"
	"github.com/amadeusitgroup/redis-operator/pkg/broker/server"
	"github.com/amadeusitgroup/redis-operator/pkg/signal"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := config.NewBrokerConfig()
	config.Init(pflag.CommandLine)

	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
	goflag.CommandLine.Parse([]string{})

	if err := run(config); err != nil {
		glog.Errorf("Redis Service Broker returns an error:%v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run(config *config.Broker) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go signal.HandleSignal(cancelFunc)

	return runWithContext(ctx, config)
}

func runWithContext(ctx context.Context, config *config.Broker) error {
	return server.Run(ctx, config)
}
