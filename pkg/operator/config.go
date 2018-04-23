package operator

import (
	"github.com/spf13/pflag"

	brockerconfig "github.com/amadeusitgroup/redis-operator/pkg/broker/config"
	"github.com/amadeusitgroup/redis-operator/pkg/config"
)

// Config contains configuration for redis-operator
type Config struct {
	Namespace      string
	KubeConfigFile string
	Master         string
	ListenAddr     string
	Redis          config.Redis
	ActivateBroker bool
	Broker         brockerconfig.Broker
}

// NewRedisOperatorConfig builds and returns a redis-operator Config
func NewRedisOperatorConfig() *Config {

	return &Config{}
}

// AddFlags add cobra flags to populate Config
func (c *Config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.Namespace, "namespace", "default", "namespace where the operator is running")
	fs.StringVar(&c.KubeConfigFile, "kubeconfig", c.KubeConfigFile, "Location of kubecfg file for access to kubernetes master service")
	fs.StringVar(&c.Master, "master", c.Master, "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&c.ListenAddr, "addr", "0.0.0.0:8086", "listen address of the http server which serves kubernetes probes and prometheus endpoints")
	fs.BoolVar(&c.ActivateBroker, "activate-broker", false, "active the broker feature in operator")
	c.Broker.Server.Init(fs)
	c.Redis.AddFlags(fs)
}
