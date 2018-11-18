package rediscluster

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	"github.com/pmorie/go-open-service-broker-client/v2"

	"k8s.io/client-go/kubernetes"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	rapi "github.com/amadeusitgroup/redis-operator/pkg/api/redis/v1"
	"github.com/amadeusitgroup/redis-operator/pkg/broker/config"
	redisv1 "github.com/amadeusitgroup/redis-operator/pkg/client/clientset/versioned"
	"github.com/amadeusitgroup/redis-operator/pkg/utils"
)

var _ broker.Interface = &BusinessLogic{}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	config      config.RedisCluster
	kubeClient  kubernetes.Interface
	redisClient redisv1.Interface

	service *v2.Service
}

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(c config.RedisCluster) (*BusinessLogic, error) {
	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.
	restConfig, err := initKubeConfig(&c)
	if err != nil {
		return nil, err
	}

	kclient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	rclient, err := redisv1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &BusinessLogic{
		config:      c,
		kubeClient:  kclient,
		redisClient: rclient,
		service:     GenerateBrokerService(),
	}, nil
}

// ValidateBrokerAPIVersion encapsulates the business logic of validating
// the OSB API version sent to the broker with every request and returns
// an error.
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#api-version-header
func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	// todo implement this method
	return nil
}

// GetCatalog encapsulates the business logic for returning the broker's
// catalog of services. Brokers must tell platforms they're integrating with
// which services they provide. GetCatalog is called when a platform makes
// initial contact with the broker to find out about that broker's services.
//
// The parameters are:
// - a RequestContext object which encapsulates:
//    - a response writer, in case fine-grained control over the response is
//      required
//    - the original http request, in case access is required (to get special
//      request headers, for example)
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#catalog-management
func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*osb.CatalogResponse, error) {
	// Your catalog business logic goes here
	response := &osb.CatalogResponse{}
	response.Services = append(response.Services, *b.service)

	return response, nil
}

// Provision encapsulates the business logic for a provision operation and
// returns a osb.ProvisionResponse or an error. Provisioning creates a new
// instance of a particular service.
//
// The parameters are:
// - a osb.ProvisionRequest created from the original http request
// - a RequestContext object which encapsulates:
//    - a response writer, in case fine-grained control over the response is
//      required
//    - the original http request, in case access is required (to get special
//      request headers, for example)
//
// Implementers should return a ProvisionResponse for a successful operation
// or an error. The APISurface handles translating ProvisionResponses or
// errors into the correct form in the http response.
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#provisioning
func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*osb.ProvisionResponse, error) {

	if !request.AcceptsIncomplete {
		return nil, fmt.Errorf("unable to provision a RedisCluster in a synchronous way")
	}

	namespace := b.config.Namespace
	clusterName := request.InstanceID
	nbMaster := int32(3)
	replica := int32(1)
	nodeResourceConf := ClusterNodeResourcesConf{}
	provisioning := provisioningSelfHosted
	userNs := ""
	if request.Context != nil {
		if val, ok := request.Context["namespace"]; ok {
			userNs = val.(string)
		}
	}
	if request.Parameters != nil {
		for key, val := range request.Parameters {
			switch parameterKey(key) {
			case clusterNameParameterKey:
				clusterName = val.(string)
			case numberOfMasterParameterKey:
				nbMaster = int32(val.(float64))
			case replicationFactorParameterKey:
				replica = int32(val.(float64))
			case memoryByNodeParameterKey:
				nodeResourceConf.MemoryRequest = val.(string)
			case cpuByNodeParameterKey:
				nodeResourceConf.CPURequest = val.(string)
			case provisioningParameterKey:
				provisioning = provisioningType(val.(string))
			default:
				glog.V(6).Infof("ignore parameter key:%s, value:%v", key, val)
			}
		}
	}

	switch provisioning {
	case provisioningUserNamespace:
		namespace = userNs
	}

	listClusters, err := b.redisClient.RedisoperatorV1().RedisClusters(namespace).List(kmetav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(listClusters.Items) > 1 {
		return nil, fmt.Errorf("redis-cluster %s/%s already exist", namespace, clusterName)
	}

	newCluster, err := NewRedisCluster(clusterName, namespace, request.InstanceID, "latest", nbMaster, replica, nodeResourceConf)
	if err != nil {
		return nil, err
	}
	_, err = b.redisClient.RedisoperatorV1().RedisClusters(namespace).Create(newCluster)
	if err != nil {
		return nil, err
	}

	response := &osb.ProvisionResponse{
		Async: true,
	}

	return response, nil
}

// Deprovision encapsulates the business logic for a deprovision operation
// and returns a osb.DeprovisionResponse or an error. Deprovisioning deletes
// an instance of a service and releases the resources associated with it.
//
// The parameters are:
// - a osb.DeprovisionRequest created from the original http request
// - a RequestContext object which encapsulates:
//    - a response writer, in case fine-grained control over the response is
//      required
//    - the original http request, in case access is required (to get special
//      request headers, for example)
//
// Implementers should return a DeprovisionResponse for a successful
// operation or an error. The APISurface handles translating
// DeprovisionResponses or errors into the correct form in the http
// response.
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#deprovisioning
func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*osb.DeprovisionResponse, error) {
	cluster, err := b.getCluster(request.InstanceID)
	if err != nil {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}
	namespace := cluster.Namespace
	deletePolicy := kmetav1.DeletePropagationForeground
	delOpts := &kmetav1.DeleteOptions{PropagationPolicy: &deletePolicy}
	if err = b.redisClient.RedisoperatorV1().RedisClusters(namespace).Delete(cluster.Name, delOpts); err != nil {
		glog.Errorf("unable to delete redis-cluster with id:%s, name:%s in namespace:%s err:%v", request.InstanceID, cluster.Name, namespace, err)

		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	return &osb.DeprovisionResponse{}, nil
}

// LastOperation encapsulates the business logic for a last operation
// request and returns a osb.LastOperationResponse or an error.
// LastOperation is called when a platform checks the status of an ongoing
// asynchronous operation on an instance of a service.
//
// The parameters are:
// - a osb.LastOperationRequest created from the original http request
// - a RequestContext object which encapsulates:
//    - a response writer, in case fine-grained control over the response is
//      required
//    - the original http request, in case access is required (to get special
//      request headers, for example)
//
// Implementers should return a LastOperationResponse for a successful
// operation or an error. The APISurface handles translating
// LastOperationResponses or errors into the correct form in the http
// response.
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#polling-last-operation
func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*osb.LastOperationResponse, error) {
	cluster, err := b.getCluster(request.InstanceID)
	if err != nil {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	response := &osb.LastOperationResponse{
		State: osb.StateInProgress,
	}

	isClusteReady := false
	if utils.GetClusterStatus(cluster) == rapi.RedisClusterOK {
		isClusteReady = true
	}

	if isClusteReady {
		response.State = osb.StateSucceeded
	}

	return response, nil
}

// Bind encapsulates the business logic for a bind operation and returns a
// osb.BindResponse or an error. Binding creates a new set of credentials for
// a consumer to use an instance of a service. Not all services are
// bindable; in order for a service to be bindable, either the service or
// the current plan associated with the instance must declare itself to be
// bindable.
//
// The parameters are:
// - a osb.BindRequest created from the original http request
// - a RequestContext object which encapsulates:
//    - a response writer, in case fine-grained control over the response is
//      required
//    - the original http request, in case access is required (to get special
//      request headers, for example)
//
// Implementers should return a BindResponse for a successful operation or
// an error. The APISurface handles translating BindResponses or errors into
// the correct form in the http response.
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#binding
func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*osb.BindResponse, error) {
	// Your bind business logic goes here
	glog.Infof("BIND rq:%#v", request)

	cluster, err := b.getCluster(request.InstanceID)
	if err != nil {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}
	credentials := make(map[string]interface{})
	credentials["service-name"] = fmt.Sprintf("%s.%s", cluster.Spec.ServiceName, cluster.Namespace)

	return &osb.BindResponse{Credentials: credentials}, nil
}

// Unbind encapsulates the business logic for an unbind operation and
// returns a osb.UnbindResponse or an error. Unbind deletes a binding and the
// resources associated with it.
//
// The parameters are:
// - a osb.UnbindRequest created from the original http request
// - a RequestContext object which encapsulates:
//    - a response writer, in case fine-grained control over the response is
//      required
//    - the original http request, in case access is required (to get special
//      request headers, for example)
//
// Implementers should return a UnbindResponse for a successful operation or
// an error. The APISurface handles translating UnbindResponses or errors
// into the correct form in the http response.
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#unbinding
func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*osb.UnbindResponse, error) {
	// Your unbind business logic goes here
	return &osb.UnbindResponse{}, nil
}

// Update encapsulates the business logic for an update operation and
// returns a osb.UpdateInstanceResponse or an error. Update updates the
// instance.
//
// The parameters are:
// - a osb.UpdateInstanceRequest created from the original http request
// - a RequestContext object which encapsulates:
//    - a response writer, in case fine-grained control over the response is
//      required
//    - the original http request, in case access is required (to get special
//      request headers, for example)
//
// Implementers should return a UpdateInstanceResponse for a successful operation or
// an error. The APISurface handles translating UpdateInstanceResponses or errors
// into the correct form in the http response.
//
// For more information, see:
//
// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#updating-a-service-instance
func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*osb.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	return &osb.UpdateInstanceResponse{}, nil
}

func (b *BusinessLogic) getCluster(instanceID string) (*rapi.RedisCluster, error) {
	rq, err := labels.NewRequirement("cluster-instance-id", selection.Equals, []string{instanceID})
	if err != nil {
		return nil, err
	}
	opts := kmetav1.ListOptions{
		LabelSelector: labels.NewSelector().Add(*rq).String(),
	}
	listClusters, err := b.redisClient.RedisoperatorV1().RedisClusters("").List(opts)
	if err != nil {
		errString := fmt.Sprintf("redis-cluster with id:%s not found, err:%v", instanceID, err)
		glog.Errorf(errString)
		return nil, fmt.Errorf(errString)
	}
	if len(listClusters.Items) != 1 {
		errString := fmt.Sprintf("redis-cluster with id:%s not found", instanceID)
		glog.Errorf(errString)

		return nil, fmt.Errorf(errString)
	}
	return &listClusters.Items[0], nil
}
