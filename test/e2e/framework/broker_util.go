package framework

import (
	"encoding/json"
	"fmt"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"

	redisclusterbroker "github.com/amadeusitgroup/redis-operator/pkg/broker/rediscluster"
)

// NewRedisClusterServiceInstance return new RedisCluster ServiceInstance
func NewRedisClusterServiceInstance(name, ns, planName string, params *redisclusterbroker.ServiceInstanceParameter) *scv1beta1.ServiceInstance {
	si := &scv1beta1.ServiceInstance{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			PlanReference: scv1beta1.PlanReference{
				ClusterServiceClassExternalName: "redis-broker-service",
				ClusterServicePlanExternalName:  planName,
			},
			Parameters: convertParametersIntoRawExtension(params),
		},
	}
	return si
}

// NewServiceInstanceParameters return new ServiceIntanceParameters
func NewServiceInstanceParameters(name, namespace string, nbMaster, replicationFactor *uint32) *redisclusterbroker.ServiceInstanceParameter {
	nbMasterInt := uint32(3)
	replicationFactorInt := uint32(1)
	if nbMaster != nil {
		nbMasterInt = *nbMaster
	}
	if replicationFactor != nil {
		replicationFactorInt = *replicationFactor
	}

	return &redisclusterbroker.ServiceInstanceParameter{
		ClusterName:       name,
		ClusterNamespace:  namespace,
		NumberOfMasters:   &nbMasterInt,
		ReplicationFactor: &replicationFactorInt,
	}
}

// HOCreateServiceInstance is an higher order func that returns the func to create a ServiceInstance
func HOCreateServiceInstance(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance) func() error {
	return func() error {
		if _, err := createInstance(c, namespace, instance); err != nil {
			Warningf("cannot create ServiceInstance %s/%s: %v", namespace, instance.Name, err)
			return err
		}
		Logf("ServiceInstance created")
		return nil
	}
}

// HOUpdateServiceInstance is an higher order func that returns the func to update a ServiceInstance
func HOUpdateServiceInstance(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance) func() error {
	return func() error {
		if _, err := updateInstance(c, namespace, instance); err != nil {
			Warningf("cannot update ServiceInstance %s/%s: %v", namespace, instance.Name, err)
			return err
		}
		Logf("ServiceInstance updated")
		return nil
	}
}

// HODeleteServiceInstance is an higher order func that returns the func to delete a ServiceInstance
func HODeleteServiceInstance(c scclientset.Interface, namespace, name string) func() error {
	return func() error {
		if err := deleteInstance(c, namespace, name); err != nil {
			Warningf("cannot delete ServiceInstance %s/%s: %v", namespace, name, err)
			return err
		}
		Logf("ServiceInstance deleted")
		return nil
	}
}

// HOIsServiceInstanceCreated is an higher order func that returns the func to know if a ServiceInstance is created properly
func HOIsServiceInstanceCreated(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance, checkStatus bool) func() error {
	return func() error {
		if err := isInstanceCreated(c, namespace, instance, checkStatus); err != nil {
			return err
		}
		Logf("ServiceInstance created properly")
		return nil
	}
}

// HOIsServiceInstanceSucceed is an higher order func that returns the func to know if a ServiceInstance is succeed properly
func HOIsServiceInstanceSucceed(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance, checkStatus bool) func() error {
	return func() error {
		if err := isInstanceSucceed(c, namespace, instance, checkStatus); err != nil {
			return err
		}
		Logf("ServiceInstance succeed properly")
		return nil
	}
}

// convertParametersIntoRawExtension converts the specified map of parameters
// into a RawExtension object that can be used in the Parameters field of
// ServiceInstanceSpec or ServiceBindingSpec.
func convertParametersIntoRawExtension(parameters *redisclusterbroker.ServiceInstanceParameter) *kruntime.RawExtension {
	marshalledParams, err := json.Marshal(parameters)
	if err != nil {
		Failf("Failed to marshal parameters %v : %v", parameters, err)
	}
	return &kruntime.RawExtension{Raw: marshalledParams}
}

// createInstance in the specified namespace
func createInstance(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance) (*scv1beta1.ServiceInstance, error) {
	return c.ServicecatalogV1beta1().ServiceInstances(namespace).Create(instance)
}

// updateInstance in the specified namespace
func updateInstance(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance) (*scv1beta1.ServiceInstance, error) {
	oldInstance, err := c.ServicecatalogV1beta1().ServiceInstances(namespace).Get(instance.Name, kmetav1.GetOptions{})
	if err != nil {
		Warningf("cannot get the service instance %s/%s: %v", namespace, instance.Name, err)
	}
	oldInstance.Spec = instance.Spec
	return c.ServicecatalogV1beta1().ServiceInstances(namespace).Update(oldInstance)
}

// updateInstance in the specified namespace
func isInstanceCreated(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance, checkStatus bool) error {
	curInstance, err := c.ServicecatalogV1beta1().ServiceInstances(namespace).Get(instance.Name, kmetav1.GetOptions{})
	if err != nil {
		Warningf("cannot get the service instance %s/%s: %v", namespace, instance.Name, err)
		return err
	}
	if checkStatus && curInstance.Status.DeprovisionStatus != scv1beta1.ServiceInstanceDeprovisionStatusRequired {
		return fmt.Errorf("Creation not done current status:%s", curInstance.Status.DeprovisionStatus)
	}
	return nil
}

// HOCreateServiceBinding is an higher order func that returns the func to create a ServiceBinding
func HOCreateServiceBinding(c scclientset.Interface, namespace string, binding *scv1beta1.ServiceBinding) func() error {
	return func() error {
		if _, err := createBinding(c, namespace, binding); err != nil {
			Warningf("cannot create ServiceBinding %s/%s: %v", namespace, binding.Name, err)
			return err
		}
		Logf("ServiceBinding created")
		return nil
	}
}

// HODeleteServiceBinding is an higher order func that returns the func to delete a ServiceBinding
func HODeleteServiceBinding(c scclientset.Interface, namespace, name string) func() error {
	return func() error {
		if err := deleteBinding(c, namespace, name); err != nil {
			Warningf("cannot delete ServiceBinding %s/%s: %v", namespace, name, err)
			return err
		}
		Logf("ServiceBinding deleted")
		return nil
	}
}

// HOIsServiceBindingCreated is an higher order func that returns the func to know if a ServiceInstance is created properly
func HOIsServiceBindingCreated(c scclientset.Interface, namespace string, binding *scv1beta1.ServiceBinding) func() error {
	return func() error {
		if err := isBindingCreated(c, namespace, binding); err != nil {
			return err
		}
		Logf("ServiceBinding created properly")
		return nil
	}
}

// isInstanceSucceed in the specified namespace
func isInstanceSucceed(c scclientset.Interface, namespace string, instance *scv1beta1.ServiceInstance, checkStatus bool) error {
	curInstance, err := c.ServicecatalogV1beta1().ServiceInstances(namespace).Get(instance.Name, kmetav1.GetOptions{})
	if err != nil {
		Warningf("cannot get the service instance %s/%s: %v", namespace, instance.Name, err)
		return err
	}

	for _, c := range curInstance.Status.Conditions {
		if c.Type == scv1beta1.ServiceInstanceConditionReady && c.Status == scv1beta1.ConditionTrue {
			return nil
		}
	}

	return fmt.Errorf("instance didn't succeed yet")
}

// deleteInstance with the specified namespace and name
func deleteInstance(c scclientset.Interface, namespace, name string) error {
	return c.ServicecatalogV1beta1().ServiceInstances(namespace).Delete(name, nil)
}

// createBinding in the specified namespace
func createBinding(c scclientset.Interface, namespace string, binding *scv1beta1.ServiceBinding) (*scv1beta1.ServiceBinding, error) {
	return c.ServicecatalogV1beta1().ServiceBindings(namespace).Create(binding)
}

// deleteBinding with the specified namespace and name
func deleteBinding(c scclientset.Interface, namespace, name string) error {
	return c.ServicecatalogV1beta1().ServiceBindings(namespace).Delete(name, nil)
}

// isBindingCreated in the specified namespace
func isBindingCreated(c scclientset.Interface, namespace string, binding *scv1beta1.ServiceBinding) error {
	_, err := c.ServicecatalogV1beta1().ServiceBindings(namespace).Get(binding.Name, kmetav1.GetOptions{})
	if err != nil {
		Warningf("cannot get the service binding %s/%s: %v", namespace, binding.Name, err)
		return err
	}

	return nil
}

// NewRedisClusterServiceBinding return new RedisCluster ServiceBinding
func NewRedisClusterServiceBinding(name, ns, instanceName string) *scv1beta1.ServiceBinding {
	si := &scv1beta1.ServiceBinding{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: scv1beta1.ServiceBindingSpec{
			ServiceInstanceRef: scv1beta1.LocalObjectReference{Name: instanceName},
		},
	}
	return si
}
