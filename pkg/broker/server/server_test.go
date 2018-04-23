package server

import (
	"net/http/httptest"
	"reflect"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"github.com/pmorie/osb-broker-lib/pkg/metrics"
	"github.com/pmorie/osb-broker-lib/pkg/rest"

	"github.com/amadeusitgroup/redis-operator/pkg/broker/config"
)

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type FakeBusinessLogic struct {
	validateBrokerAPIVersion func(version string) error
	getCatalog               func(c *broker.RequestContext) (*osb.CatalogResponse, error)
	provision                func(request *osb.ProvisionRequest, c *broker.RequestContext) (*osb.ProvisionResponse, error)
	deprovision              func(request *osb.DeprovisionRequest, c *broker.RequestContext) (*osb.DeprovisionResponse, error)
	lastOperation            func(request *osb.LastOperationRequest, c *broker.RequestContext) (*osb.LastOperationResponse, error)
	bind                     func(request *osb.BindRequest, c *broker.RequestContext) (*osb.BindResponse, error)
	unbind                   func(request *osb.UnbindRequest, c *broker.RequestContext) (*osb.UnbindResponse, error)
	update                   func(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*osb.UpdateInstanceResponse, error)
}

var _ broker.Interface = &FakeBusinessLogic{}

func (b *FakeBusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return b.validateBrokerAPIVersion(version)
}
func (b *FakeBusinessLogic) GetCatalog(c *broker.RequestContext) (*osb.CatalogResponse, error) {
	return b.getCatalog(c)
}
func (b *FakeBusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*osb.ProvisionResponse, error) {
	return b.provision(request, c)
}
func (b *FakeBusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*osb.DeprovisionResponse, error) {
	return b.deprovision(request, c)
}
func (b *FakeBusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*osb.LastOperationResponse, error) {
	return b.lastOperation(request, c)
}
func (b *FakeBusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*osb.BindResponse, error) {
	return b.bind(request, c)
}
func (b *FakeBusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*osb.UnbindResponse, error) {
	return b.unbind(request, c)
}
func (b *FakeBusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*osb.UpdateInstanceResponse, error) {
	return b.update(request, c)
}

func TestGetCatalog(t *testing.T) {
	expectedResponse := &osb.CatalogResponse{Services: []osb.Service{
		{
			Name: "foo",
		},
	}}

	api := &rest.APISurface{
		Broker: &FakeBusinessLogic{
			getCatalog: func(c *broker.RequestContext) (*osb.CatalogResponse, error) {
				return expectedResponse, nil
			},
			validateBrokerAPIVersion: func(version string) error { return nil },
		},
		Metrics: metrics.New(),
	}

	srvConfig := config.NewBrokerConfig()

	s := New(&srvConfig.Server, api)

	fs := httptest.NewServer(s.router)
	defer fs.Close()

	config := osb.DefaultClientConfiguration()
	config.URL = fs.URL

	client, err := osb.NewClient(config)
	if err != nil {
		t.Error(err)
	}

	actualResponse, err := client.GetCatalog()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(actualResponse, expectedResponse) {
		t.Errorf("Unexpected response\n\nExpected: %#+v\n\nGot: %#+v", expectedResponse, actualResponse)
	}
}
