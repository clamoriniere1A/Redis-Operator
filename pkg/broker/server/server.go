package server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"

	"github.com/pmorie/osb-broker-lib/pkg/rest"

	"github.com/amadeusitgroup/redis-operator/pkg/broker/config"
)

// Server service broker server
type Server struct {
	config *config.Server
	// router is a mux.Router that registers the handlers for the different OSB
	// API operations.
	router http.Handler
}

// New returns new instance of the service broker server
func New(config *config.Server, api *rest.APISurface) *Server {
	router := GetHandler(api)

	return &Server{
		config: config,
		router: router,
	}
}

// GetHandler used to register broker handler function on new Router
func GetHandler(api *rest.APISurface) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/v2/catalog", api.GetCatalogHandler).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}/last_operation", api.LastOperationHandler).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}", api.ProvisionHandler).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}", api.DeprovisionHandler).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{instance_id}", api.UpdateHandler).Methods("PATCH")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", api.BindHandler).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", api.UnbindHandler).Methods("DELETE")
	return router
}

// Run creates the HTTP handler and begins to listen on the specified address.
func (s *Server) Run(ctx context.Context, addr string) error {
	listenAndServe := func(srv *http.Server) error {
		return srv.ListenAndServe()
	}
	return s.run(ctx, addr, listenAndServe)
}

// RunTLS creates the HTTPS handler based on the certifications that were passed
// and begins to listen on the specified address.
func (s *Server) RunTLS(ctx context.Context, addr string, cert string, key string) error {
	var decodedCert, decodedKey []byte
	var tlsCert tls.Certificate
	var err error
	decodedCert, err = base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return err
	}
	decodedKey, err = base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	tlsCert, err = tls.X509KeyPair(decodedCert, decodedKey)
	if err != nil {
		return err
	}
	listenAndServe := func(srv *http.Server) error {
		srv.TLSConfig = new(tls.Config)
		srv.TLSConfig.Certificates = []tls.Certificate{tlsCert}
		return srv.ListenAndServeTLS("", "")
	}
	return s.run(ctx, addr, listenAndServe)
}

func (s *Server) run(ctx context.Context, addr string, listenAndServe func(srv *http.Server) error) error {
	glog.Infof("Starting server on %s\n", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if srv.Shutdown(c) != nil {
			srv.Close()
		}
	}()
	return listenAndServe(srv)
}
