package config

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
)

// Broker service broker configuration
type Broker struct {
	Server       Server
	Rediscluster RedisCluster
	PromRegistry *prom.Registry
}

// NewBrokerConfig returns new Broker configuration instance
func NewBrokerConfig() *Broker {
	return &Broker{}
}

// Init used to add command line parameter for the service broker server
func (b *Broker) Init(fs *pflag.FlagSet) {
	b.Server.Init(fs)
	b.Rediscluster.Init(fs)

}

// Server server configuration
type Server struct {
	Port    int
	TLSCert string
	TLSKey  string
}

// Init used to add command line parameter for the service broker server
func (s *Server) Init(fs *pflag.FlagSet) {
	fs.IntVar(&s.Port, "broker-port", 8005, "use '--port' option to specify the port for broker to listen on")
	fs.StringVar(&s.TLSCert, "broker-tlsCert", "", "base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used. If '--tlsCert' is not used, then TLS will not be used.")
	fs.StringVar(&s.TLSKey, "broker-tlsKey", "", "base-64 encoded PEM block to use as the private key matching the TLS certificate. If '--tlsKey' is used, then '--tlsCert' must also be used")
}

// RedisCluster redis-cluster service broker configuration
type RedisCluster struct {
	KubeConfigFile string
	Master         string
	Namespace      string
}

// Init used to add command line parameters for the service broker redis-cluster logic
func (r *RedisCluster) Init(fs *pflag.FlagSet) {
	// Nothing todo for now.
	fs.StringVar(&r.KubeConfigFile, "kubeconfig", r.KubeConfigFile, "Location of kubecfg file for access to kubernetes master service")
	fs.StringVar(&r.Master, "master", r.Master, "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
