package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pmorie/osb-broker-lib/pkg/metrics"
	"github.com/pmorie/osb-broker-lib/pkg/rest"

	"github.com/amadeusitgroup/redis-operator/pkg/broker/config"
	"github.com/amadeusitgroup/redis-operator/pkg/broker/rediscluster"
	"github.com/golang/glog"
)

// Run the service broker
func Run(ctx context.Context, config *config.Broker) error {
	if (config.Server.TLSCert != "" || config.Server.TLSKey != "") &&
		(config.Server.TLSCert == "" || config.Server.TLSKey == "") {
		fmt.Println("To use TLS, both --tlsCert and --tlsKey must be used")
		return nil
	}

	addr := ":" + strconv.Itoa(config.Server.Port)

	osbMetrics := metrics.New()
	config.PromRegistry.MustRegister(osbMetrics)

	businessLogic, err := rediscluster.NewBusinessLogic(config.Rediscluster)
	if err != nil {
		return err
	}

	api, err := rest.NewAPISurface(businessLogic, osbMetrics)
	if err != nil {
		return err
	}

	s := New(&config.Server, api)

	glog.Infof("Starting broker!")

	if config.Server.TLSCert == "" && config.Server.TLSKey == "" {
		err = s.Run(ctx, addr)
	} else {
		err = s.RunTLS(ctx, addr, config.Server.TLSCert, config.Server.TLSKey)
	}
	return err
}

// Broker represent the Broker instance
type Broker struct {
	config *config.Broker
	server *Server
}

// NewBroker returns new Broker instance
func NewBroker(config *config.Broker) (*Broker, error) {
	if (config.Server.TLSCert != "" || config.Server.TLSKey != "") &&
		(config.Server.TLSCert == "" || config.Server.TLSKey == "") {
		return nil, fmt.Errorf("To use TLS, both --broker-tlsCert and --broker-tlsKey must be used")
	}

	osbMetrics := metrics.New()
	config.PromRegistry.MustRegister(osbMetrics)

	businessLogic, err := rediscluster.NewBusinessLogic(config.Rediscluster)
	if err != nil {
		return nil, err
	}

	api, err := rest.NewAPISurface(businessLogic, osbMetrics)
	if err != nil {
		return nil, err
	}

	srv := New(&config.Server, api)

	broker := &Broker{
		config: config,
		server: srv,
	}

	return broker, nil
}

// Run use to start tke Broker
func (b *Broker) Run(ctx context.Context) error {
	addr := ":" + strconv.Itoa(b.config.Server.Port)
	glog.Infof("Starting broker!")
	var err error
	if b.config.Server.TLSCert == "" && b.config.Server.TLSKey == "" {
		err = b.server.Run(ctx, addr)
	} else {
		err = b.server.RunTLS(ctx, addr, b.config.Server.TLSCert, b.config.Server.TLSKey)
	}

	return err
}
